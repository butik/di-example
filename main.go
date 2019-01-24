package main

import (
	"encoding/json"
	"github.com/globalsign/mgo"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	dbname = "test"
	productsCollection = "products"
)

type ProductDatastore interface {
	Find(id int) Product
	Save(product Product)
}

var _ ProductDatastore = &ProductDatastoreMongo{}
type ProductDatastoreMongo struct {
	session *mgo.Session
}

func (p *ProductDatastoreMongo) Save(product Product) {
	p.session.DB(dbname).C(productsCollection).Insert(product)
}

func (p *ProductDatastoreMongo) Find(id int) Product {
	var result Product
	p.session.DB(dbname).C(productsCollection).FindId(id).One(&result)

	return result
}

func NewProductDatastoreMongo(session *mgo.Session) *ProductDatastoreMongo {
	return &ProductDatastoreMongo{
		session: session,
	}
}

type Product struct {
	ID int `bson:"_id"`
	Name string `bson:"name"`
}

type FeedService struct {
	datastore ProductDatastore
	search SearchService
}

func NewFeedService(datastore ProductDatastore, search SearchService) *FeedService {
	return &FeedService{
		datastore: datastore,
		search: search,
	}
}

type SearchService interface {
	SearchProductIDs() []int
}

var _ SearchService = &SearchServiceElastic{}
type SearchServiceElastic struct {

}

func (*SearchServiceElastic) SearchProductIDs() []int {
	return []int {1, 2, 3}
}

func NewSearchServiceElastic() *SearchServiceElastic {
	return &SearchServiceElastic{}
}

func (f *FeedService) GenerateProductFeed() []Product {
	productIDs := f.search.SearchProductIDs()

	result := make([]Product, 0)
	for _, productID := range productIDs {
		product := f.datastore.Find(productID)
		result = append(result, product)
	}

	return result
}

func feedhandler(feedService *FeedService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		products := feedService.GenerateProductFeed()

		json.NewEncoder(w).Encode(products)
	}
}

func loadFixtures(datastore ProductDatastore) {
	datastore.Save(Product{ID: 1, Name: "Product1"})
	datastore.Save(Product{ID: 2, Name: "Product2"})
	datastore.Save(Product{ID: 3, Name: "Product3"})
}

func main1() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	productDatastore := NewProductDatastoreMongo(session)

	// clean db and load fixtures
	session.DB(dbname).DropDatabase()
	loadFixtures(productDatastore)

	searchService := NewSearchServiceElastic()
	feedService := NewFeedService(productDatastore, searchService)

	r := mux.NewRouter()

	r.Path("/feed").Methods("GET").Handler(feedhandler(feedService))

	srv := &http.Server{
		Addr: ":8080",
		Handler: r,
	}
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		panic(err)
	}
}
