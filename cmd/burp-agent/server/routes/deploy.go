package routes

import (
	"burp/cmd/burp-agent/server"
	"burp/cmd/burp-agent/server/mimes"
	responses2 "burp/cmd/burp-agent/server/responses"
	"burp/internal/burpy"
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
		logger := responses2.Logger(ctx)
		if ctx.ContentType() != "multipart/form-data" {
			responses2.InvalidPayload.Reply(ctx)
			return
		}
		form, err := ctx.MultipartForm()
		if err != nil {
			responses2.HandleErr(ctx, err)
			return
		}
		files := form.File["package[]"]
		if len(files) == 0 {
			responses2.InvalidPayload.Reply(ctx)
			return
		}
		var configBytes []byte
		var pkg *uploadedFile
		for _, file := range files {
			file := file

			contentType := file.Header.Get("Content-Type")
			if !utils.AnyMatchString(ACCEPTED_FILE_MIMETYPES, contentType) {
				logger.Error().Str("Content-Type", contentType).Msg("Invalid Payload")
				responses2.InvalidPayload.Reply(ctx)
				return
			}
			f, err := file.Open()
			if err != nil {
				responses2.HandleErr(ctx, err)
				return
			}
			bytes, err := io.ReadAll(f)
			if err != nil {
				responses2.HandleErr(ctx, err)
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
					responses2.InvalidPayload.Reply(ctx)
					return
				}
				mime := mimetype.Detect(bytes)
				if !mime.Is(mimes.TAR_MIMETYPE) {
					logger.Error().Str("Mime", mime.String()).Msg("Invalid Payload")
					responses2.InvalidPayload.Reply(ctx)
					return
				}
				pkg = utils.Ptr(uploadedFile{Name: fileName, Contents: bytes})
			}
			if contentType == mimes.TOML_MIMETYPE {
				configBytes = bytes
			}
		}
		logger.Info().Msg("Starting server-side stream...")
		responses2.AddSseHeaders(ctx)

		channel := utils.Ptr(make(chan any, 10))
		go func() {
			defer close(*channel)

			// IMPT: All deployments should be synchronous to  prevent an existential crisis
			// that doesn't exist, but still to be safe.
			responses2.ChannelSend(channel, responses2.CreateChannelOk("Waiting for deployment agent..."))
			logger.Info().Msg("Waiting for deployment agent...")

			lock.Lock()
			defer lock.Unlock()

			var burp services.Burp
			if err = toml.Unmarshal(configBytes, &burp); err != nil {
				logger.Err(err).Msg("Failed to parse TOML file into Burp services")
				responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to parse TOML file into Burp services", err.Error()))
				return
			}

			if pkg != nil {
				tarName := fmt.Sprint(burp.Service.Name, "_includes.tar.gz")
				if pkg.Name != tarName {
					responses2.ChannelSend(channel, responses2.ErrorResponse{Error: "Invalid uploaded package.", Code: http.StatusBadRequest})
					return
				}
				logger.Info().Msg("Unpacking uploaded files...")
				responses2.ChannelSend(channel, responses2.CreateChannelOk("Unpacking uploaded files..."))
				dir := filepath.Join(burpy.TemporaryFilesFolder, burp.Service.Name)
				if err = fileutils.MkdirParent(dir); err != nil {
					responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to create temporary files folder", err.Error()))
					return
				}
				buffer := bytes.NewReader(pkg.Contents)
				if err = extract.Archive(context.TODO(), buffer, dir, nil); err != nil {
					responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to unpack uploaded package", err.Error()))
					return
				}
				responses2.ChannelSend(channel, responses2.CreateChannelOk("Validating checksums of unpacked files..."))
				var hashes []services.HashedInclude
				metaFileBytes, err := os.ReadFile(filepath.Join(dir, "meta.json"))
				if err != nil {
					responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to read metadata of unpacked files", err.Error()))
					return
				}
				if err = json.Unmarshal(metaFileBytes, &hashes); err != nil {
					responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to read metadata of unpacked files", err.Error()))
					return
				}
				for _, include := range hashes {
					file := filepath.Base(include.Target)
					file = filepath.Join(dir, "pkg", file)
					f, err := fileutils.Open(file)
					if err != nil {
						responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to read metadata of unpacked files", err.Error()))
						return
					}
					hash := sha256.New()
					if _, err := io.Copy(hash, f); err != nil {
						responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to read metadata of unpacked files", err.Error()))
						return
					}
					fileutils.Close(f)
					checksum := hex.EncodeToString(hash.Sum(nil))
					if checksum != include.Hash {
						responses2.ChannelSend(channel, responses2.ErrorResponse{Error: "Checksum of files does not match.", Code: http.StatusBadRequest})
						return
					}
					target := filepath.Clean(include.Target)
					target = filepath.Join(".burpy", "home", target)

					responses2.ChannelSend(channel, responses2.CreateChannelOk("File "+include.Source+" passed checksum, copying to "+target))
					if _, err = fileutils.Copy(file, target); err != nil {
						responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to copy file to destination", err.Error()))
						return
					}
				}
			}
			logger.Info().Msg("Starting build process...")
			responses2.ChannelSend(channel, responses2.CreateChannelOk("Starting build process..."))
			burpy.Deploy(channel, &burp)
			responses2.ChannelSend(channel, responses2.CreateChannelOk("Cleaning all stages..."))
			if err := burpy.Clear(&burp); err != nil {
				responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to clean all stages", err.Error()))
				return
			}
		}()

		responses2.Stream(ctx, channel)
	})
})
