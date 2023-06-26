package routes

import (
	"burp/cmd/burp-agent/server"
	"burp/cmd/burp-agent/server/limiter"
	"burp/cmd/burp-agent/server/mimes"
	responses "burp/cmd/burp-agent/server/responses"
	"burp/internal/burp"
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
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// _
// PUT /application: You can use this route to deploy an application.
//
// Requires: [Content-Type=multipart,File=[package[],burp.toml,application/toml]]
// Optional: [Content-Type=multipart,File=[package[],(service)_includes.tar.gz,application/gzip]]
// Returns: sse-stream
// _
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
		if len(files) == 0 {
			responses.InvalidPayload.Reply(ctx)
			return
		}
		var configBytes []byte
		var environments []string
		var pkg *uploadedFile
		for _, file := range files {
			file := file

			contentType := file.Header.Get("Content-Type")
			if !utils.AnyMatchString(AcceptedFileMimetypes, contentType) {
				logger.Error().Str("Content-Type", contentType).Msg("Invalid Payload")
				responses.InvalidPayload.Reply(ctx)
				return
			}
			f, err := file.Open()
			if err != nil {
				responses.HandleErr(ctx, err)
				return
			}
			bits, err := io.ReadAll(f)
			if err != nil {
				responses.HandleErr(ctx, err)
				return
			}
			fileName := filepath.Base(file.Filename)
			// IMPT: Always double-check the tar filetype since it could be disguised before we unarchive it.
			// We can skip TOML.
			//
			// Also, the correct Content-Type is GZIP, but the actual file is a TAR, which is why
			// we have two different mimetypes in this check.
			if contentType == mimes.GZIP_MIMETYPE {
				if !utils.HasSuffixStr(fileName, ".tar.gz") {
					logger.Error().Str("file", fileName).Msg("Invalid Payload")
					responses.InvalidPayload.Reply(ctx)
					return
				}
				mime := mimetype.Detect(bits)
				if !mime.Is(mimes.TAR_MIMETYPE) {
					logger.Error().Str("Mime", mime.String()).Msg("Invalid Payload")
					responses.InvalidPayload.Reply(ctx)
					return
				}
				pkg = utils.Ptr(uploadedFile{Name: fileName, Contents: bits})
			}
			if contentType == mimes.TOML_MIMETYPE {
				configBytes = bits
			}
			if contentType == mimes.TEXT_MIMETYPE && fileName == ".env" {
				environments = burp.EnvironmentReadBuffer(bytes.NewReader(bits))
			}
		}
		logger.Info().Msg("Spawning server-side stream...")
		responses.AddSseHeaders(ctx)
		responses.Stream(ctx, func(context context.Context, channel *chan any) {
			limiter.Await(channel, logger)
			defer limiter.GlobalAgentLock.Unlock()

			var application burp.Application
			if ok := application.From(configBytes, logger, channel); !ok {
				return
			}

			if pkg != nil {
				tarName := fmt.Sprint(application.Service.Name, "_includes.tar.gz")
				if pkg.Name != tarName {
					responses.ChannelSend(channel, responses.ErrorResponse{Error: "Invalid uploaded package.", Code: http.StatusBadRequest})
					return
				}
				logger.Info().Msg("Unpacking uploaded files...")
				responses.Message(channel, "Unpacking uploaded files...")
				dir := filepath.Join(burp.TemporaryFilesFolder, application.Service.Name)
				if err = fileutils.MkdirParent(dir); err != nil {
					responses.Error(channel, "Failed to create temporary files folder", err)
					return
				}
				buffer := bytes.NewReader(pkg.Contents)
				if err = extract.Archive(context, buffer, dir, nil); err != nil {
					responses.Error(channel, "Failed to unpack uploaded package", err)
					return
				}
				responses.Message(channel, "Validating checksums of unpacked files...")
				var hashes []burp.HashedInclude
				metaFileBytes, err := os.ReadFile(filepath.Join(dir, "meta.json"))
				if err != nil {
					responses.Error(channel, "Failed to read metadata of unpacked files", err)
					return
				}
				if err = json.Unmarshal(metaFileBytes, &hashes); err != nil {
					responses.Error(channel, "Failed to read metadata of unpacked files", err)
					return
				}
				for _, include := range hashes {
					file := filepath.Base(include.Target)
					file = filepath.Join(dir, "pkg", file)
					f, err := fileutils.Open(file)
					if err != nil {
						responses.Error(channel, "Failed to read metadata of unpacked files", err)
						return
					}
					hash := sha256.New()
					if _, err := io.Copy(hash, f); err != nil {
						responses.Error(channel, "Failed to read metadata of unpacked files", err)
						return
					}
					fileutils.Close(f)
					checksum := hex.EncodeToString(hash.Sum(nil))
					if checksum != include.Hash {
						responses.ChannelSend(channel, responses.ErrorResponse{Error: "Checksum of files does not match.", Code: http.StatusBadRequest})
						return
					}
					target := filepath.Clean(include.Target)
					target = filepath.Join(burp.UnpackedFilesFolder, target)

					responses.Message(channel, "File ", include.Source, " passed checksum, copying to ", target)
					if _, err = fileutils.Copy(file, target); err != nil {
						responses.Error(channel, "Failed to copy file to destination", err)
						return
					}
				}
			}
			logger.Info().Msg("Starting build process...")
			responses.Message(channel, "Starting build process...")
			defer func() {
				if err := application.CleanRemnants(); err != nil {
					logger.Err(err).Msg("Failed to clean remnants.")
					return
				}
				logger.Debug().Msg("Cleaned remnants.")
			}()
			application.Deploy(context, channel, environments)
		})
	})
})

var AcceptedFileMimetypes = []string{
	mimes.GZIP_MIMETYPE,
	mimes.TOML_MIMETYPE,
	mimes.TEXT_MIMETYPE,
}

type uploadedFile struct {
	Name     string
	Contents []byte
}
