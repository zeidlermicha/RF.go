package Regression

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"time"
)

type Feature interface {
	string | float64 | int
}

type Forest[T Feature] struct {
	Trees []*Tree[T]
}

func BuildForest[T Feature](inputs [][]T, labels []float64, treesAmount, samplesAmount, selectedFeatureAmount int) *Forest[T] {
	NumWorkers := runtime.NumCPU()
	forest := &Forest[T]{}
	forest.Trees = make([]*Tree[T], treesAmount)
	prog_counter := 0
	mutex := &sync.Mutex{}
	s := make(chan bool, NumWorkers)
	for i := 0; i < treesAmount; i++ {
		s <- true

		go func(x int) {
			defer func() { <-s }()
			fmt.Printf(">> %v buiding %vth tree...\n", time.Now(), x)
			forest.Trees[x] = BuildTree(inputs, labels, samplesAmount, selectedFeatureAmount)
			//fmt.Printf("<< %v the %vth tree is done.\n",time.Now(), x)
			mutex.Lock()
			prog_counter += 1
			fmt.Printf("%v tranning progress %.0f%%\n", time.Now(), float64(prog_counter)/float64(treesAmount)*100)
			mutex.Unlock()
		}(i)
	}
	for i := 0; i < NumWorkers; i++ {
		s <- true
	}

	fmt.Println("all done.")
	return forest
}

func DefaultForest[T Feature](inputs [][]T, labels []float64, treesAmount int) *Forest[T] {
	m := int(math.Sqrt(float64(len(inputs[0]))))
	n := int(math.Sqrt(float64(len(inputs))))
	return BuildForest[T](inputs, labels, treesAmount, n, m)
}

func (self *Forest[T]) Predicate(input []T) float64 {
	total := 0.0
	for i := 0; i < len(self.Trees); i++ {
		total += PredicateTree(self.Trees[i], input)
	}
	avg := total / float64(len(self.Trees))
	return avg
}

func DumpForest[T Feature](forest *Forest[T], fileName string) {
	out_f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		panic("failed to create " + fileName)
	}
	defer out_f.Close()
	encoder := json.NewEncoder(out_f)
	encoder.Encode(forest)
}

func LoadForest[T Feature](fileName string) *Forest[T] {
	in_f, err := os.Open(fileName)
	if err != nil {
		panic("failed to open " + fileName)
	}
	defer in_f.Close()
	decoder := json.NewDecoder(in_f)
	forest := &Forest[T]{}
	decoder.Decode(forest)
	return forest
}
