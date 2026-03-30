package subscriber_test

import (
	"errors"
	"fmt"

	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/alexfalkowski/go-health/v2/subscriber"
)

func ExampleNewObserver() {
	sub := subscriber.NewSubscriber([]string{"db"})
	ob := subscriber.NewObserver([]string{"db"}, sub)

	errPing := errors.New("ping failed")
	sub.Send(probe.NewTick("db", errPing))
	sub.Close()
	ob.Wait()

	fmt.Println(errors.Is(ob.Error(), errPing))
	fmt.Println(errors.Is(ob.Errors()["db"], errPing))
	// Output:
	// true
	// true
}
