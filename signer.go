package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	c := 0
	var results1 [100]string
	var results2 [100]string
	var wg sync.WaitGroup
	m := make(chan string, 1)
	for dataRaw := range in {
		data, ok := dataRaw.(int)
		if !ok {
			fmt.Println("SingleHash: Unsupported data type")
			return
		}
		wg.Add(2)
		dataStr := strconv.Itoa(data)
		go func(i int) {
			defer wg.Done()
			m <- ""
			md5ResultCh := DataSignerMd5(dataStr)
			<-m
			results1[i] = DataSignerCrc32(md5ResultCh)
		}(c)
		go func(i int) {
			defer wg.Done()
			results2[i] = DataSignerCrc32(dataStr)

		}(c)
		c++
	}
	wg.Wait()
	for i := 0; i < c; i++ {
		out <- results2[i] + "~" + results1[i]
	}

}

func MultiHash(in, out chan interface{}) {
	var hashResults [100][6]chan string
	c := 0
	var wg sync.WaitGroup //Waitgroup для ожидания завершения горутин

	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			fmt.Println("MultiHash: Unsuported data type")
			return
		}
		wg.Add(1)
		for j := 0; j < 6; j++ {
			hashResults[c][j] = make(chan string, 1)
		}
		for th := 0; th < 6; th++ {
			go func(j int, i int, data string) {
				hashResults[i][j] <- DataSignerCrc32(strconv.Itoa(j) + data)

			}(th, c, data)
			go func(i int) {
				defer wg.Done()
				result := ""
				for j := 0; j < 6; j++ {
					result += <-hashResults[i][j]
				}
				out <- result
			}(c)
		}
		c++
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string

	for result := range in {
		resultStr, ok := result.(string)
		if !ok {
			fmt.Println("CombineResults: Unsupported data type")
			return
		}
		results = append(results, resultStr)
	}
	sort.Strings(results)
	combinedResult := strings.Join(results, "_")

	out <- combinedResult
}

func ExecutePipeline(jobs ...job) {
	//Создаем каналы для каждого задания
	var chans []chan interface{}

	for range jobs {
		chans = append(chans, make(chan interface{}))
	}

	var wg sync.WaitGroup

	for i, j := range jobs {
		in := chans[i]
		var out chan interface{}
		if i == len(jobs)-1 {
			out = make(chan interface{}) //Если задание последнее  создаем канал для вывода результата
		} else {
			out = chans[i+1] //Если нет, то передавать результат будем в канал следующего задания
		}

		wg.Add(1)
		go func(jobFunc job, in, out chan interface{}) { //Запускаем горутину для выполнения задания
			defer wg.Done()
			jobFunc(in, out)
			close(out)
		}(j, in, out)
	}
	wg.Wait()
}
