package routes

import (
	"burp/internal/burper"
	"burp/internal/burpy"
	"burp/internal/server"
	"burp/internal/server/mimes"
	"burp/internal/server/responses"
	"burp/internal/services"
	"burp/pkg/fileutils"
	"burp/pkg/utils"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/codeclysm/extract/v3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var ACCEPTED_FILE_MIMETYPES = []string{
	mimes.GZIP_MIMETYPE,
	mimes.TOML_MIMETYPE,
}

type uploadedFile struct {
	Name     string
	Contents []byte
}

var lock = sync.Mutex{}

var _ = server.Add(func(app *gin.Engine) {
	app.PUT("/application", func(ctx *gin.Context) {
		logger := responses.Logger(ctx)
		if ctx.ContentType() != "multipart/form-data" {
			responses.InvalidPayload.Reply(ctx)
			return
		}
		form, err := ctx.MultipartForm()
		if err != nil {
			responses.HandleErr(ctx, err)
			return
		}
		files := form.File["package[]"]
		var burp []byte
		var pkg *uploadedFile
		for _, file := range files {
			file := file

			contentType := file.Header.Get("Content-Type")
			if !utils.AnyMatchString(ACCEPTED_FILE_MIMETYPES, contentType) {
				logger.Error().Str("Content-Type", contentType).Msg("Invalid Payload")
				responses.InvalidPayload.Reply(ctx)
				return
			}
			f, err := file.Open()
			if err != nil {
				responses.HandleErr(ctx, err)
				return
			}
			bytes, err := io.ReadAll(f)
			if err != nil {
				responses.HandleErr(ctx, err)
				return
			}
			// IMPT: Always double-check the tar filetype since it could be disguised before we unarchive it.
			// We can skip TOML.
			//
			// Also, the correct Content-Type is GZIP, but the actual file is a TAR, which is why
			// we have two different mimetypes in this check.
			if contentType == mimes.GZIP_MIMETYPE {
				fileName := filepath.Base(file.Filename)
				if !utils.HasSuffixStr(fileName, ".tar.gz") {
					logger.Error().Str("file", fileName).Msg("Invalid Payload")
					responses.InvalidPayload.Reply(ctx)
					return
				}
				mime := mimetype.Detect(bytes)
				if !mime.Is(mimes.TAR_MIMETYPE) {
					logger.Error().Str("Mime", mime.String()).Msg("Invalid Payload")
					responses.InvalidPayload.Reply(ctx)
					return
				}
				pkg = utils.Ptr(uploadedFile{Name: fileName, Contents: bytes})
			}
			if contentType == mimes.TOML_MIMETYPE {
				burp = bytes
			}
		}
		logger.Info().Msg("Starting server-side stream...")
		ctx.Writer.Header().Set("Content-Type", "text/event-stream")
		ctx.Writer.Header().Set("Cache-Control", "no-cache")
		ctx.Writer.Header().Set("Connection", "keep-alive")
		ctx.Writer.Header().Set("Transfer-Encoding", "chunked")

		channel := utils.Ptr(make(chan any, 10))

		// IMPT: All deployments should be synchronous to  prevent an existential crisis
		// that doesn't exist, but still to be safe.
		responses.ChannelSend(channel, responses.CreateChannelOk("Waiting for deployment agent..."))
		logger.Info().Msg("Waiting for deployment agent...")

		lock.TryLock()

		go func() {
			defer lock.Unlock()

			tree, err := burper.FromBytes(burp)
			if err != nil {
				logger.Info().Err(err).Msg("Failed to parse TOML file into Burp tree")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to parse TOML file into Burp tree", err.Error()))
				return
			}

			var burp services.Burp
			if err = toml.Unmarshal(tree.Bytes(), &burp); err != nil {
				logger.Info().Err(err).Msg("Failed to parse TOML file into Burp services")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to parse TOML file into Burp services", err.Error()))
				return
			}

			if pkg != nil {
				tarName := fmt.Sprint(burp.Service.Name, "_includes.tar.gz")
				if pkg.Name != tarName {
					responses.ChannelSend(channel, responses.ErrorResponse{Error: "Invalid uploaded package.", Code: http.StatusBadRequest})
					return
				}
				logger.Info().Msg("Unpacking uploaded files...")
				responses.ChannelSend(channel, responses.CreateChannelOk("Unpacking uploaded files..."))
				dir := filepath.Join(burpy.TemporaryFilesFolder, burp.Service.Name)
				if err = fileutils.MkdirParent(dir); err != nil {
					responses.ChannelSend(channel, responses.CreateChannelError("Failed to create temporary files folder", err.Error()))
					return
				}
				buffer := bytes.NewBuffer(pkg.Contents)
				if err = extract.Archive(context.TODO(), buffer, dir, nil); err != nil {
					responses.ChannelSend(channel, responses.CreateChannelError("Failed to unpack uploaded package", err.Error()))
					return
				}
				responses.ChannelSend(channel, responses.CreateChannelOk("Validating checksums of unpacked files..."))
				var hashes []services.HashedInclude
				metaFileBytes, err := os.ReadFile(filepath.Join(dir, "meta.json"))
				if err != nil {
					responses.ChannelSend(channel, responses.CreateChannelError("Failed to read metadata of unpacked files", err.Error()))
					return
				}
				if err = json.Unmarshal(metaFileBytes, &hashes); err != nil {
					responses.ChannelSend(channel, responses.CreateChannelError("Failed to read metadata of unpacked files", err.Error()))
					return
				}
				for _, include := range hashes {
					file := filepath.Base(include.Target)
					file = filepath.Join(dir, "pkg", file)
					f, err := fileutils.Open(file)
					if err != nil {
						responses.ChannelSend(channel, responses.CreateChannelError("Failed to read metadata of unpacked files", err.Error()))
						return
					}
					hash := sha256.New()
					if _, err := io.Copy(hash, f); err != nil {
						responses.ChannelSend(channel, responses.CreateChannelError("Failed to read metadata of unpacked files", err.Error()))
						return
					}
					fileutils.Close(f)
					checksum := hex.EncodeToString(hash.Sum(nil))
					if checksum != include.Hash {
						responses.ChannelSend(channel, responses.ErrorResponse{Error: "Checksum of files does not match.", Code: http.StatusBadRequest})
						return
					}
					target := filepath.Clean(include.Target)
					target = filepath.Join(".burpy", "home", target)

					responses.ChannelSend(channel, responses.CreateChannelOk("File "+include.Source+" passed checksum, copying to "+target))
					if _, err = fileutils.Copy(file, target); err != nil {
						responses.ChannelSend(channel, responses.CreateChannelError("Failed to copy file to destination", err.Error()))
						return
					}
				}
			}
			logger.Info().Msg("Starting build process...")
			responses.ChannelSend(channel, responses.CreateChannelOk("Starting build process..."))
			burpy.Deploy(channel, &burp)
			defer close(*channel)
		}()

		ctx.Stream(func(w io.Writer) bool {
			if msg, ok := <-*channel; ok {
				log.Info().Any("data", msg).Msg("Received stream message")
				ctx.SSEvent("data", msg)
				return true
			}
			return false
		})
	})
})
