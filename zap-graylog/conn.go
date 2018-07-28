package zapGL

import (
	"log"
	"net"
	"strconv"

	"github.com/Broadroad/gpool"
)

func giveConnPool(conf Conf) *gpool.Pool {

	var pool *gpool.Pool

	if conf.GrayLogConnType == GrayConnTCP {
		pool = makeTCPConnPool(conf)
	}

	if pool == nil && conf.PanicOnFail {
		panic("Connection attempt to graylog server failed")
	}

	return pool
}

func makeTCPConnPool(conf Conf) *gpool.Pool {
	var addr = conf.GrayLogHostname + ":" + strconv.Itoa(conf.GrayLogConnPort)

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)

	if err != nil {
		log.Printf("Unable to resolve graylog server addresss\nAddress : %s\nError : %s", addr, err)
		return nil
	}

	tcpAddrStr := tcpAddr.String()

	factory := func() (net.Conn, error) { return net.Dial("tcp", tcpAddrStr) }

	poolConfig := &gpool.PoolConfig{
		InitCap: 16,
		MaxCap:  64,
		Factory: factory,
	}

	// create a new conn pool
	p, err := gpool.NewGPool(poolConfig)

	if err != nil {
		log.Printf("Unable to connect to graylog server \nAddress : %s\nError : %s", addr, err)
		return nil
	}

	return &p
}
