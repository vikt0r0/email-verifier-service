// apiserver/apiserver.go
package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	emailverifier "github.com/vikt0r0/email-verifier"
)

var defaultStopTimeout = time.Second * 30

type APIServer struct {
	addr string
}

func NewAPIServer(addr string) (*APIServer, error) {
	if addr == "" {
		return nil, errors.New("addr cannot be blank")
	}

	return &APIServer{
		addr: addr,
	}, nil
}

// Start starts a server with a stop channel
func (s *APIServer) Start(stop <-chan struct{}) error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.router(),
	}

	go func() {
		logrus.WithField("addr", srv.Addr).Info("starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), defaultStopTimeout)
	defer cancel()

	logrus.WithField("timeout", defaultStopTimeout).Info("stopping server")
	return srv.Shutdown(ctx)
}

func (s *APIServer) router() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/", s.defaultRoute)
	return router
}

var (
	verifier = emailverifier.
		NewVerifierWithEmailAndName(verifierFromEmail, verifierHelloName).
		EnableSMTPCheck()
)

func (s *APIServer) defaultRoute(w http.ResponseWriter, r *http.Request) {

	type Response struct {
		Result       emailverifier.Result
		Error        bool
		ErrorMessage string
	}

	var email = r.URL.Query().Get("email")

	ret, err := verifier.Verify(email)

	var errorString string

	if err != nil {
		errorString = fmt.Sprintf("verify email address failed, error is: %s", err)
	}

	if !ret.Syntax.Valid {
		errorString = "email address syntax is invalid"
	}

	if email == "" {
		errorString = "no or empty email GET parameter specified"
	}

	var responseStruct = Response{
		Result:       *ret,
		Error:        errorString != "",
		ErrorMessage: errorString,
	}

	b, err := json.MarshalIndent(responseStruct, "", "  ")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(b))
}
