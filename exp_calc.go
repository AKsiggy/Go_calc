package main

import (
	"errors"
	"fmt"
	"sync"
)

var res float64

type Calculator interface {
	Add(a float64, b float64, wg *sync.WaitGroup) float64
	Subtract(a float64, b float64, wg *sync.WaitGroup) float64
	Multiply(a float64, b float64, wg *sync.WaitGroup) float64
	Divide(a float64, b float64, wg *sync.WaitGroup) (float64, error)
}

type SimpleCalculator struct {
	mu sync.Mutex
}

func (c *SimpleCalculator) Add(a float64, b float64, wg *sync.WaitGroup) float64 {
	c.mu.Lock()
	res = a + b
	defer c.mu.Unlock()
	wg.Done()
	return res
}

func (c *SimpleCalculator) Subtract(a float64, b float64, wg *sync.WaitGroup) float64 {
	c.mu.Lock()
	res = a - b
	defer c.mu.Unlock()
	wg.Done()
	return res
}

func (c *SimpleCalculator) Multiply(a float64, b float64, wg *sync.WaitGroup) float64 {
	c.mu.Lock()
	res = a * b
	defer c.mu.Unlock()
	wg.Done()
	return res
}

func (c *SimpleCalculator) Divide(a float64, b float64, wg *sync.WaitGroup) (float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	wg.Done()
	return a / b, nil
}

type Operation struct {
	Op    string
	A, B  float64
	Reply chan float64
	Err   chan error
}

func main() {
	calculator := &SimpleCalculator{}
	operations := make(chan Operation)
	var wg sync.WaitGroup

	go func() {
		for op := range operations {
			var result float64
			var err error
			wg.Add(1)
			switch op.Op {
			case "add":
				result = calculator.Add(op.A, op.B, &wg)
			case "subtract":
				result = calculator.Subtract(op.A, op.B, &wg)
			case "multiply":
				result = calculator.Multiply(op.A, op.B, &wg)
			case "divide":
				result, err = calculator.Divide(op.A, op.B, &wg)
			default:
				err = errors.New("unknown operation")
			}

			op.Reply <- result
			op.Err <- err
		}
	}()
	wg.Wait()
	handleUserInput(operations)

	close(operations)
}

func handleUserInput(operations chan Operation) {
	var opType string
	var a, b float64

	for {
		fmt.Println("Enter operation (add, subtract, multiply, divide) or 'exit' to quit:")
		fmt.Scanln(&opType)
		if opType == "exit" {
			break
		}

		fmt.Println("Enter the first number:")
		fmt.Scanln(&a)
		fmt.Println("Enter the second number:")
		fmt.Scanln(&b)

		reply := make(chan float64)
		errChan := make(chan error)

		operations <- Operation{Op: opType, A: a, B: b, Reply: reply, Err: errChan}
		result := <-reply
		err := <-errChan

		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("Result: %.2f\n", result)
		}
	}
}
