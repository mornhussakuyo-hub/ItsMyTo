package main

import webview "github.com/webview/webview_go"

func runDesktop(url string, debug bool) error {
	window := webview.New(debug)
	defer window.Destroy()

	window.SetTitle(appName)
	window.SetSize(1180, 760, webview.HintNone)
	window.Navigate(url)
	window.Run()
	return nil
}
