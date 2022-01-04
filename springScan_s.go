package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	numberTasks                []string
	the_returned_result_is_200 []string
	list_of_errors             []string
	t                          = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	src_file          string
	target_file       string
	des_file          string
	routineCountTotal int
	url               string
	targetPaths       []string
)

func title() {
	fmt.Println(`
  ▄████  ▒█████  
 ██▒ ▀█▒▒██▒  ██▒
▒██░▄▄▄░▒██░  ██▒
░▓█  ██▓▒██   ██░
░▒▓███▀▒░ ████▓▒░
 ░▒   ▒ ░ ▒░▒░▒░ 
  ░   ░   ░ ▒ ▒░ 
░ ░   ░ ░ ░ ░ ▒  
      ░     ░ ░
	 Here is springScan.
`)
}
func main() {
	flag.StringVar(&src_file, "s", "spring.txt", "字典文件")
	flag.StringVar(&target_file, "f", "url.txt", "目标网站文件")
	flag.StringVar(&url, "u", "", "目标url")
	flag.StringVar(&des_file, "d", "result.txt", "结果文件")
	flag.IntVar(&routineCountTotal, "t", 40, "线程数量{默认为40}")
	flag.Parse()
	title()
	file, err := os.Open(src_file)
	if err != nil {
		fmt.Println("打开文件时候出错")
	}
	defer func() {
		file.Close()
	}()
	n := bufio.NewScanner(file)
	for n.Scan() {
		data := n.Text()
		numberTasks = append(numberTasks, data)

	}

	targetPaths, err = OpenTargetFile(target_file)
	if err != nil {
		fmt.Printf("open target file failed, msg:%v\n", err)
	}

	fmt.Printf("numberTasks: %v\n", numberTasks)
	client = &http.Client{
		Transport: t,
		Timeout:   20 * time.Second,
	}
	beg := time.Now()
	wg := &sync.WaitGroup{}
	tasks := make(chan string)
	results := make(chan string)
	go func() {
		for result := range results {
			if result == "" {
				close(results)
			} else if strings.Contains(result, "200") {
				fmt.Printf("result loop:%v\n", result)
				the_returned_result_is_200 = append(the_returned_result_is_200, result)
			} else {
				list_of_errors = append(list_of_errors, result)
			}
		}
	}()

	for _, path := range targetPaths {
		for i := 0; i < routineCountTotal; i++ {
			wg.Add(1)
			go worker(wg, tasks, results, path)
		}
		for _, task := range numberTasks {
			tasks <- task
		}
	}

	tasks <- ""
	wg.Wait()
	results <- ""
	fmt.Println("\033[33m+++++++++++++++++++请求成功的++++++++++++++++++++++")

	file_1, err := os.OpenFile(des_file, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	defer file_1.Close()
	write_1 := bufio.NewWriter(file_1)
	for _, v := range the_returned_result_is_200 {
		fmt.Println(v)
		write_1.WriteString(v + "\n")
	}
	write_1.Flush()
	fmt.Println("发生了", len(list_of_errors), "个失败")
	fmt.Printf("time consumed: %fs\n", time.Now().Sub(beg).Seconds())
	fmt.Println("具体接口用法请参考:https://github.com/LandGrey/SpringBootVulExploit")

}

func worker(group *sync.WaitGroup, tasks chan string, result chan string, path string) {
	for task := range tasks {
		if task == "" {
			close(tasks)
		} else {
			respBody, err := NumberQueryRequest(task, path)
			if err != nil {
				fmt.Printf("error occurred in NumberQueryRequest: %s\n", task)
				result <- err.Error()
			} else {
				result <- respBody
			}
		}
	}
	group.Done()
}

var client *http.Client

func NumberQueryRequest(keyword string, path string) (body string, err error) {
	url := fmt.Sprintf("%s%s", path, keyword)
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "构造请求出错", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
	resp, err := client.Get(url)
	if err != nil {
		return "发送请求出错", err
	}
	return_value := resp.StatusCode
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	body = "url:" + url + " || " + "返回值:" + strconv.Itoa(return_value)
	return body, nil

}

func OpenTargetFile(targetFileName string) (targetTasks []string, err error) {
	file, err := os.Open(targetFileName)
	if err != nil {
		fmt.Println("打开文件时候出错")
	}
	defer func() {
		file.Close()
	}()
	n := bufio.NewScanner(file)
	for n.Scan() {
		data := n.Text()
		targetTasks = append(targetTasks, data)

	}
	fmt.Printf("targetTasks: %v\n", targetTasks)
	return
}