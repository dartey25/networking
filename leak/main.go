package main

func worker(ch chan any) {
	for range ch {
	}
}

func main() {
	leakage()
}

func leakage() {
	ch := make(chan any)
	go worker(ch)
}
