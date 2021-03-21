package cmd

import (
	"context"
	"io"
	"os"
	"fmt"
	"github.com/flano-yuki/t2q2t/config"
	"github.com/flano-yuki/t2q2t/lib"
	quic "github.com/lucas-clemente/quic-go"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"net"
)

var s2qCmd = &cobra.Command{
	Use:   "s2q",
	Short: "Stream to quic",
	Long: `Stream to quic
  t2q2t s2q <forward Addr>  

  go run ./t2q2t.go s2q 127.0.0.1:2022:`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		to := args[0]

		err := runs2q(to)
		if err != nil {
			fmt.Printf("[Error] %s\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(s2qCmd)
}

func runs2q(to string) error {
	toTcpAddr, err := net.ResolveTCPAddr("tcp4", to)
	if err != nil {
		panic(err)
	}


	tlsConf := config.GenerateClientTLSConfig()
	quicConf := config.GenerateClientQUICConfig()
	var sess quic.Session = nil

	// TODO
	// and, if connection has closed
	if sess == nil {
		sess, err = quic.DialAddr(toTcpAddr.String(), tlsConf, quicConf)
		if err != nil {
			return err
		}
//		fmt.Printf("Connect QUIC to: %s \n", toTcpAddr.String())
	}

	// TODO error handling
	s2qHandleConn(os.Stdin, os.Stdout, sess)
	return nil
}

func s2qHandleConn(reader io.Reader, writer io.Writer, sess quic.Session) error {
	var stream quic.Stream
	stream, err := sess.OpenStreamSync(context.Background())
	if err != nil {
		panic(err)
	}

	eg := errgroup.Group{}
	eg.Go(func() error { return util.S2qRelay(reader, stream) })
	eg.Go(func() error { return util.Q2sRelay(stream, writer) })

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
