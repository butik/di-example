package main

import (
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/gorilla/mux"
	"github.com/sarulabs/di"
	"net/http"
)

const (
	diMongoPool = "mongo-pool"
	diMongoSession = "mongo-session"
	diProductDatastore = "product-datastore"
	diSearchService = "search-service"
	diFeedService = "feed-service"
)

func WithContainer(app di.Container) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return di.HTTPMiddleware(next.ServeHTTP, app, func(msg string) {
			fmt.Printf("container error: %v\n", msg)
		})
	}
}

func feedhandlerDi(w http.ResponseWriter, r *http.Request) {
		feedService := di.Get(r, diFeedService).(*FeedService)
		products := feedService.GenerateProductFeed()

		json.NewEncoder(w).Encode(products)
}

func main() {
	builder, _ := di.NewBuilder()

	builder.Add(di.Def{
		Scope: di.App,
		Name: diMongoPool,
		Build: func(ctn di.Container) (interface{}, error) {
			return mgo.Dial("localhost")
		},
		Close: func(obj interface{}) error {
			obj.(*mgo.Session).Close()
			return nil
		},
	})
	builder.Add(di.Def{
		Scope: di.Request,
		Name: diMongoSession,
		Build: func(ctn di.Container) (interface{}, error) {
			session := ctn.Get(diMongoPool).(*mgo.Session).Copy()
			return session, nil
		},
		Close: func(obj interface{}) error {
			obj.(*mgo.Session).Close()
			return nil
		},
	})
	builder.Add(di.Def{
		Scope: di.Request,
		Name: diProductDatastore,
		Build: func(ctn di.Container) (interface{}, error) {
			session := ctn.Get(diMongoSession).(*mgo.Session)

			return NewProductDatastoreMongo(session), nil
		},
	})
	builder.Add(di.Def{
		Scope: di.Request,
		Name: diSearchService,
		Build: func(ctn di.Container) (interface{}, error) {
			return NewSearchServiceElastic(), nil
		},
	})
	builder.Add(di.Def{
		Scope: di.Request,
		Name: diFeedService,
		Build: func(ctn di.Container) (interface{}, error) {
			return NewFeedService(
				ctn.Get(diProductDatastore).(ProductDatastore),
				ctn.Get(diSearchService).(SearchService)), nil
		},
	})

	r := mux.NewRouter()

	app := builder.Build()
	defer app.Delete()

	r.Use(WithContainer(app))

	r.Path("/feed").Methods("GET").HandlerFunc(feedhandlerDi)

	srv := &http.Server{
		Addr: ":8080",
		Handler: r,
	}
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		panic(err)
	}
}
