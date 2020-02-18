/*Package atexit lets you define handlers when the program exits.

Add handlers using Register.
You must call atexit.Exit to get the handler invoked (and then terminate the program).

This package also provides replacements to log.Fatal, log.Fatalf and log.Fatalln.

Example:

    package main

    import (
        "fmt"

        "github.com/tebeka/atexit"
    )

    func handler() {
        fmt.Println("Exiting")
    }

    func main() {
            atexit.Register(handler)
            atexit.Exit(0)
    }
*/
package atexit

import (
	"fmt"
	"log"
	"os"
	"sync"
)

const (
	// Version is package version
	Version = "0.3.0"
)

type HandlerId uint

func (h *HandlerId) Get() HandlerId {
	defer func() {
		*h++
	}()

	return *h
}

var handlers = map[HandlerId]func(){}
var handlersLock sync.RWMutex
var once sync.Once
var nextFuncId HandlerId

func runHandler(handler func()) {
	defer func() {
		if err := recover(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "error: atexit handler error:", err)
		}
	}()

	handler()
}

func executeHandlers() {
	handlersLock.RLock()
	defer handlersLock.RUnlock()
	for _, handler := range handlers {
		runHandler(handler)
	}
}

func runHandlers() {
	once.Do(executeHandlers)
}

// Exit runs all the atexit handlers and then terminates the program using
// os.Exit(code)
func Exit(code int) {
	runHandlers()
	os.Exit(code)
}

// Fatal runs all the atexit handler then calls log.Fatal (which will terminate
// the program)
func Fatal(v ...interface{}) {
	runHandlers()
	log.Fatal(v...)
}

// Fatalf runs all the atexit handler then calls log.Fatalf (which will
// terminate the program)
func Fatalf(format string, v ...interface{}) {
	runHandlers()
	log.Fatalf(format, v...)
}

// Fatalln runs all the atexit handler then calls log.Fatalln (which will
// terminate the program)
func Fatalln(v ...interface{}) {
	runHandlers()
	log.Fatalln(v...)
}

// Register adds a handler and return the handler's id,
// call atexit.Exit to invoke all handlers.
func Register(handler func()) HandlerId {
	handlersLock.Lock()
	defer handlersLock.Unlock()

	currentId := nextFuncId.Get()
	handlers[currentId] = handler

	return currentId
}

// Unregister remove a handler
func Unregister(id HandlerId) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	delete(handlers, id)
}
