# FyshSaver

A screensaver app built using Fyne which augments the FyneDesk desktop and FyshOS system.

[![FyshSaver preview](https://img.youtube.com/vi/GjSFY2-y4KU/0.jpg)](https://www.youtube.com/watch?v=GjSFY2-y4KU)

## Usage

This project can be called as a library as follows:

```go
save := saver.NewScreenSaver(func() {
	log.Println("Exited")
})
save.Lock = true

save.ShowWindow()
```

You can also run this screensaver with a demo app in the
`cmd/fyshsaver` folder:

```bash
    cd cmd/fyshsaver
    go run .
```
