package save

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

const AliasLength = 6

//go:generate go run github.com/vektra/mockery/v2@v2 --name=URLSaver --with-expecterf
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		var req Request

		err := render.DecodeJSON(request.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(writer, request, resp.Error("request body is empty"))
			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(writer, request, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors

			if errors.As(err, &validateErr) {
				log.Error("invalid request", sl.Err(err))
				render.JSON(writer, request, resp.ValidatorError(validateErr))
			} else {
				log.Error("unexpected error during validation", sl.Err(err))
				render.JSON(writer, request, resp.Error("internal server error"))
			}

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(AliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrUrlExist) {
			maxAttempts := 2
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				alias = random.NewRandomString(AliasLength)
				id, err = urlSaver.SaveURL(req.URL, alias)

				if err == nil {
					log.Info("url added", slog.Int64("id", id), slog.String("alias", alias))
					responseOK(writer, request, alias)
					return
				}

				if !errors.Is(err, storage.ErrUrlExist) {
					break
				}

				log.Warn("alias collision", slog.Int("attempt", attempt), slog.String("alias", alias))
			}

			log.Info("url already exist", slog.String("url", req.URL), slog.Int("status_code", http.StatusConflict))

			render.JSON(writer, request, resp.Error("url already exist"))

			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(writer, request, resp.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		responseOK(writer, request, alias)
	}
}

func responseOK(writer http.ResponseWriter, request *http.Request, alias string) {
	render.JSON(writer, request, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
