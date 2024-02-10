// dsts prints status updates to standard output, as expected by i3bar.
//
// See <https://i3wm.org/docs/i3bar-protocol.html>.
package main

import (
	"context"
	"os"

	"github.com/c032/dsts"
	dststime "github.com/c032/dsts/time"
)

func must1(err error) {
	if err != nil {
		panic(err)
	}
}

func must2[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

func main() {
	ctx := context.Background()

	cli := must2(dsts.NewCLI(ctx, os.Stdout))

	tn := &dststime.Notifier{}

	must1(cli.AddNotifier(tn))
	must1(cli.AddStatusLineBlocks(&tn.StatusUnix, &tn.StatusDateTime))

	err := cli.Run()
	if err != nil {
		panic(err)
	}
}
