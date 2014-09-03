package main

func main() {
	app := NewApp()
	app.NewWindow()

	<-app.ir.Done
}
