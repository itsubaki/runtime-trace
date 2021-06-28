package main

import (
	"context"
	"log"
	"os"
	"runtime/trace"
	"sync"
)

type Matrix [][]complex128

func New(v ...[]complex128) Matrix {
	out := make(Matrix, len(v))
	for i := 0; i < len(v); i++ {
		out[i] = v[i]
	}

	return out
}

func (m Matrix) Dimension() (int, int) {
	return len(m), len(m[0])
}

func (m Matrix) Apply(n Matrix) Matrix {
	ctx, task := trace.NewTask(context.Background(), "Apply")
	defer task.End()

	p, _ := m.Dimension()
	a, b := n.Dimension()

	wg := sync.WaitGroup{}
	out := make(Matrix, a)
	for i := 0; i < a; i++ {
		wg.Add(1)
		go func(ctx context.Context, i int, out *Matrix) {
			defer wg.Done()
			defer trace.StartRegion(ctx, "apply").End()

			v := make([]complex128, b)
			for j := 0; j < b; j++ {
				c := complex(0, 0)
				for k := 0; k < p; k++ {
					c = c + n[i][k]*m[k][j]
				}

				v = append(v, c)
			}

			(*out)[i] = v
		}(ctx, i, &out)
	}

	wg.Wait()
	return out
}

func _main() {
	m := New(
		[]complex128{0, 0, 0, 1},
		[]complex128{1, 0, 0, 0},
		[]complex128{0, 1, 0, 0},
		[]complex128{0, 0, 1, 0},
	)

	m.Apply(m)
}

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("create file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("close file: %v", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		log.Fatalf("start trace: %v", err)
	}
	defer trace.Stop()

	_main()
}
