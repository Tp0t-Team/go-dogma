package dogma

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"reflect"
)

type Context = echo.Context

type Method struct {
	Name   string `json:"name"`
	Method string `json:"method"`
}

type Server struct {
	router *echo.Group
	desc   map[reflect.Type]Method
	logger *log.Logger
}

func New(router *echo.Group, desc map[reflect.Type]Method, logger *log.Logger) *Server {
	return &Server{
		router: router,
		desc:   desc,
		logger: logger,
	}
}

func typedParse[T any](data []byte) (*T, error) {
	ret := new(T)
	err := json.Unmarshal(data, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type ResultBase struct {
	Message string `json:"message"`
}

func HandleRestFunc[T func(Context, PU, PC) (R, error), PU any, PC any, R any](s *Server, handle T) {
	typ := reflect.TypeOf(new(T))
	desc, ok := s.desc[typ]
	if !ok {
		log.Panicln(fmt.Sprintf("no such api for type: %s", typ.String()))
	}
	var fn func(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route = nil
	switch desc.Method {
	case "GET":
		fn = s.router.GET
		break
	case "POST":
		fn = s.router.POST
		break
	case "PUT":
		fn = s.router.PUT
		break
	case "DELETE":
		fn = s.router.DELETE
		break
	case "PATCH":
		fn = s.router.PATCH
		break
	}
	if fn == nil {
		return
	}
	fn(fmt.Sprintf("/%s", desc.Name), func(ctx echo.Context) error {
		vars := map[string]string{}
		for _, name := range ctx.ParamNames() {
			vars[name] = ctx.Param(name)
		}
		varsRaw, err := json.Marshal(vars)
		if err != nil {
			if s.logger != nil {
				s.logger.Println("[dogma]", err)
			}
			return ctx.String(http.StatusBadRequest, "")
		}
		urlParam, err := typedParse[PU](varsRaw)
		if err != nil {
			if s.logger != nil {
				s.logger.Println("[dogma]", err)
			}
			return ctx.String(http.StatusBadRequest, "")
		}
		commonParam := new(PC)
		if desc.Method == http.MethodPost {
			commonParam := new(PC)
			err := (&echo.DefaultBinder{}).BindBody(ctx, commonParam)
			if err != nil {
				if s.logger != nil {
					s.logger.Println("[dogma]", err)
				}
				return ctx.String(http.StatusBadRequest, "")
			}
		}
		ret, err := handle(ctx, *urlParam, *commonParam)
		result := struct {
			ResultBase
			R
		}{}
		if err != nil {
			result.Message = err.Error()
		} else {
			result.R = ret
		}
		return ctx.JSON(http.StatusOK, ret)
	})
}
