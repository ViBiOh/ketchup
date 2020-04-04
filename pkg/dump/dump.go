package dump

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/request"
)

// Handler for dump request. Should be use with net/http
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := dumpRequest(r)

		logger.Info("Dump of request\n%s", value)

		if _, err := w.Write([]byte(value)); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}

func dumpRequest(r *http.Request) string {
	var headers bytes.Buffer
	for key, value := range r.Header {
		headers.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(value, ",")))
	}

	var params bytes.Buffer
	for key, value := range r.URL.Query() {
		headers.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(value, ",")))
	}

	var form bytes.Buffer
	if err := r.ParseForm(); err != nil {
		form.WriteString(err.Error())
	} else {
		for key, value := range r.PostForm {
			form.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(value, ",")))
		}
	}

	body, err := request.ReadBodyRequest(r)
	if err != nil {
		logger.Error("%s", err)
	}

	var outputPattern bytes.Buffer
	outputPattern.WriteString("%s %s\n")
	outputData := []interface{}{
		r.Method,
		r.URL.Path,
	}

	if headers.Len() != 0 {
		outputPattern.WriteString("Headers\n%s\n")
		outputData = append(outputData, headers.String())
	}

	if params.Len() != 0 {
		outputPattern.WriteString("Params\n%s\n")
		outputData = append(outputData, params.String())
	}

	if form.Len() != 0 {
		outputPattern.WriteString("Form\n%s\n")
		outputData = append(outputData, form.String())
	}

	if len(body) != 0 {
		outputPattern.WriteString("Body\n%s\n")
		outputData = append(outputData, body)
	}

	return fmt.Sprintf(outputPattern.String(), outputData...)
}
