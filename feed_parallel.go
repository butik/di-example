package main

import (
	"github.com/sarulabs/di"
	"sync"
)

type indexedUnit struct {
	idx     int
	productID int
}

type indexedResult struct {
	idx     int
	product Product
}

func consumer(reqCnt di.Container, work chan indexedUnit, result chan indexedResult, done *sync.WaitGroup) {
	consumerCnt, _ := reqCnt.Parent().SubContainer()
	defer consumerCnt.Delete()

	defer done.Done()
	productDatastoreRef := consumerCnt.Get(diProductDatastore)

	productDatastore := productDatastoreRef.(ProductDatastore)

	for job := range work {
		product := productDatastore.Find(job.productID)

		result <- indexedResult{
			idx:     job.idx,
			product: product,
		}
	}
}

const productQueueSize = 10
const productConsumers = 10
func generateFeedParallel(ctn di.Container, productIDs []int) {
	queue := make(chan indexedUnit, productQueueSize)
	resultQ := make(chan indexedResult, productQueueSize)

	var wg sync.WaitGroup

	for i := 0; i < productConsumers; i++ {
		wg.Add(1)
		go consumer(ctn, queue, resultQ, &wg)
	}

	for idx, productID := range productIDs {
		queue <- indexedUnit{idx: idx, productID: productID}
	}
	close(queue)

	go func() {
		wg.Wait()
		close(resultQ)
	}()

	result := make([]Product, len(productIDs))
	for productRes := range resultQ {
		result[productRes.idx] = productRes.product
	}
}
