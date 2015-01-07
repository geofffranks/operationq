package operationq_test

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/operationq"
	"github.com/pivotal-golang/operationq/fake_operationq"
)

var _ = Describe("Queue", func() {
	var queue operationq.Queue

	BeforeEach(func() {
		queue = operationq.NewQueue()
	})

	Describe("Push", func() {
		var operations []*fake_operationq.FakeOperation

		JustBeforeEach(func(done Done) {
			for _, o := range operations {
				queue.Push(o)
			}

			close(done)
		})

		Context("when there are no current operations", func() {
			BeforeEach(func() {
				operations = []*fake_operationq.FakeOperation{operationWithKey("key-1")}
			})

			It("executes the operation", func() {
				Eventually(operations[0].ExecuteCallCount).Should(Equal(1))
			})
		})

		Context("when operations have different keys", func() {
			var k1 *fake_operationq.FakeOperation
			var k2 *fake_operationq.FakeOperation
			var wait *sync.WaitGroup

			BeforeEach(func() {
				k1 = operationWithKey("k1")
				k2 = operationWithKey("k2")
				wait = new(sync.WaitGroup)
				wait.Add(1)
				k1.ExecuteStub = func() {
					defer GinkgoRecover()
					wait.Done()
					wait.Wait()
				}

				wait.Add(1)
				k2.ExecuteStub = func() {
					defer GinkgoRecover()
					Ω(k1.ExecuteCallCount()).Should(Equal(1))
					wait.Done()
					wait.Wait()
				}

				operations = []*fake_operationq.FakeOperation{k1, k2}
			})

			It("runs them in parallel", func(done Done) {
				wait.Wait()
				close(done)
			})
		})

		Context("when operations have the same key", func() {
			var k1op1 *fake_operationq.FakeOperation
			var k1op2 *fake_operationq.FakeOperation
			var out chan string

			BeforeEach(func() {
				k1op1 = operationWithKey("k1")
				k1op2 = operationWithKey("k1")
				out = make(chan string, 2)

				k1op1.ExecuteStub = func() {
					time.Sleep(time.Millisecond)
					out <- "op1"
				}

				k1op2.ExecuteStub = func() {
					out <- "op2"
				}

				operations = []*fake_operationq.FakeOperation{k1op1, k1op2}
			})

			It("runs them in order", func(done Done) {
				Ω(<-out).Should(Equal("op1"))
				Ω(<-out).Should(Equal("op2"))
				close(done)
			})
		})
	})
})

func operationWithKey(key string) *fake_operationq.FakeOperation {
	o := new(fake_operationq.FakeOperation)
	o.KeyReturns(key)
	return o
}
