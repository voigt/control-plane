package main

import (
	"log"
	"net/http"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"

	"github.com/pkg/errors"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

type config struct {
	Address               string `envconfig:"default=127.0.0.1:3000"`
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	// TODO: Get values from config
	certHeaderParser := authentication.NewHeaderParser("", "", "", "", "")
	authContextMiddleware := authentication.NewAuthenticationContextMiddleware(certHeaderParser)

	tokenResolver := api.NewTokenResolver()
	certificateResolver := api.NewCertificateResolver()
	resolver := api.Resolver{TokenResolver: tokenResolver, CertificateResolver: certificateResolver}

	gqlCfg := gqlschema.Config{
		Resolvers: &resolver,
	}
	executableSchema := gqlschema.NewExecutableSchema(gqlCfg)

	log.Printf("Registering endpoint on %s...", cfg.APIEndpoint)
	router := mux.NewRouter()
	router.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	router.HandleFunc(cfg.APIEndpoint, handler.GraphQL(executableSchema))

	router.Use(authContextMiddleware.PropagateAuthentication)

	http.Handle("/", router)

	log.Printf("Listening on %s...", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, nil); err != nil {
		panic(err)
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}