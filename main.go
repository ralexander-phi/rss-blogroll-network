package main

func main() {
	a := NewAnalysis()
	defer a.Close()

	a.Analyze()
	a.Visualize()
}
