package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os"
	"log"
	"io"
	"strings"
	//"net/http"
	"time"
	"strconv"
	"github.com/fatih/color"
)

const (
	sshPortEnv  = "SSH_PORT"
	httpPortEnv = "PORT"

	defaultSshPort  = "2022"
	defaultHttpPort = "3000"
)

func handler(conn net.Conn, gm *GameManager, config *ssh.ServerConfig) {
	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		fmt.Println("Failed to handshake with new client")
		return
	}
	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			fmt.Println("could not accept channel.")
			return
		}

		// TODO: Remove this -- only temporary while we launch on HN
		//
		// To see how many concurrent users are online
		fmt.Printf("Player joined. Current stats: %d users, %d games\n",
			gm.SessionCount(), gm.GameCount())

		// Reject all out of band requests accept for the unix defaults, pty-req and
		// shell.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				switch req.Type {
				case "pty-req":
					req.Reply(true, nil)
					continue
				case "shell":
					req.Reply(true, nil)
					continue
				}
				req.Reply(false, nil)
			}
		}(requests)

		fmt.Printf(" sshConn.User : %s\n", sshConn.User())

		for i:=0;i<len(tmpPlayer);i++{
			if(sshConn.User() == tmpPlayer[i].Name){
				fmt.Println("find Same ID")
				gm.HandleSavedChannel(channel, sshConn.User(),&tmpPlayer[i])
			}
		}
		//gm.HandleNewChannel(channel, sshConn.User())
	}
}

func port(env, def string) string {
	port := os.Getenv(env)
	if port == "" {
		port = def
	}

	return fmt.Sprintf(":%s", port)
}

type SendData struct{
		tmpID string
		tmpDirection string
		tmpMarker string
		tmpPosX string
		tmpPosY string
		tmpScore string
		tmpColor string
		checker bool
		Trail []PlayerTrailSegment
}
var(
	totalUser = 3
	sendData = [3]SendData{}

	tmpPlayer = []Player{}

	inChecker bool
)



/*	tmpID string
	tmpDirection string
	tmpMarker string
	tmpPosX string
	tmpPosY string
	tmpScore string
	tmpColor string
)*/

func ConnHandler(conn net.Conn, gm *GameManager) {
	fmt.Println("connhandler")
   recvBuf := make([]byte, 819200)
   for {
      n, err := conn.Read(recvBuf)
			//fmt.Println("%s",recvBuf)
      if nil != err {
         if io.EOF == err {
            log.Println(err);
            return
         }
         log.Println(err);
         return
      }
      if 0 < n {
         data := recvBuf[:n]
 //        log.Println(string(data))
				tmp := strings.Split(string(data),",")


				var trailList []PlayerTrailSegment
				var tmpPlayerPos Position
				var tmpTrailPos Position
				var tmpMark rune
				var tmpintcolor color.Attribute

				tmpID := tmp[0]
				floatColor ,_ := strconv.ParseFloat(tmp[1],64)
				Direction,_ := strconv.ParseFloat(tmp[2],64)
				Marker ,_ := strconv.ParseFloat(tmp[3],64)
				tmpPosX ,_ := strconv.ParseFloat(tmp[4],64)
				tmpPosY ,_ := strconv.ParseFloat(tmp[5],64)
				tmpScore ,_ := strconv.ParseFloat(tmp[6],64)


				tmpPlayerPos.X = tmpPosX
				tmpPlayerPos.Y = tmpPosY

				tmpDirection := PlayerDirection(Direction)
				tmpMarker := rune(Marker)
				tmpColor := color.Attribute(floatColor)

				for j := 7; j < len(tmp)-3; j+=4{
					mark,_ := strconv.ParseFloat(tmp[j],64)
					posx,_ := strconv.ParseFloat(tmp[j+1],64)
					posy,_ := strconv.ParseFloat(tmp[j+2],64)
					intcolor,_ := strconv.ParseFloat(tmp[j+3],64)
					//fmt.Printf("trailList : %d, \n",trailList)
					//fmt.Printf("mark : %d, posx : %d, posy : %d, color : %d\n",mark,posx,posy,intcolor)
					tmpTrailPos.X = posx
					tmpTrailPos.Y = posy
					tmpMark = rune(mark)
					tmpintcolor = color.Attribute(intcolor)
					trailList=append(trailList, PlayerTrailSegment{tmpMark,tmpTrailPos,tmpintcolor})
				}
				checker := false

				for i:=0;i<len(tmpPlayer);i++ {
					if tmpPlayer[i].Name == tmpID{
						checker =true
						tmpPlayer[i].Direction = tmpDirection
						tmpPlayer[i].Marker = tmpMarker
						tmpPlayer[i].score = tmpScore
						tmpPlayer[i].Pos = &tmpPlayerPos
						tmpPlayer[i].Trail = trailList
						break;
					}
				}

				if !checker{
					tmpPlayer = append(tmpPlayer, Player{
						Name : tmpID,
						Direction: tmpDirection,
						Marker: tmpMarker,
						Color: tmpColor,
						Pos:  &tmpPlayerPos,
						score: tmpScore,
						Trail : trailList,
						})
				}
				inChecker = true

			}

				} ///totalUser

  }

func main() {
	sshPort := port(sshPortEnv, defaultSshPort)
	httpPort := port(httpPortEnv, defaultHttpPort)
	//sendData := [3]SendData{}

	// Everyone can login!
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		panic("Failed to load private key")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
	}

	config.AddHostKey(private)

	// Create the GameManager
	gm := NewGameManager()
	//gm.SendServerNewGame()
	//fmt.Println("make new game manager")
	go func(){
		fmt.Println("data arrive func")
		l, err := net.Listen("tcp", "0.0.0.0:8000")
		if nil != err {
		 	log.Println(err);
		}
		defer l.Close()

		for {
		 conn, err := l.Accept()
		 if nil != err {
					log.Println(err);
					continue
		 		}
		 	defer conn.Close()
			fmt.Println("data arrive")
		 	go ConnHandler(conn,gm)
		}

}()

	fmt.Printf(
		"Listening on port %s for SSH and port %s for HTTP...\n",
		sshPort,
		httpPort,
	)

/*	go func() {
		panic(http.ListenAndServe(httpPort, http.FileServer(http.Dir("./static/"))))
	}()*/

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	go func(){
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0%s", sshPort))
	if err != nil {
		panic("failed to listen for connection")
	}

	for {
		nConn, err := listener.Accept()
		if err != nil {
			panic("failed to accept incoming connection")
		}
		startTime := time.Now()

		fmt.Printf("input time : %s\n",startTime)
		go handler(nConn, gm, config)
	}
}()

	go func(){
		for {
			if inChecker{
				//fmt.Println("main Name : " ,tmpPlayer[0].Name )
				//fmt.Println("main direction : ",tmpPlayer[0].Direction)/*
				/*fmt.Println("main marker : ",tmpPlayer[0].Marker)
				fmt.Println("main color : ",tmpPlayer[0].Color)
				fmt.Println("main pos : ",tmpPlayer[0].Pos )
				fmt.Println("main score : ",tmpPlayer[0].score )
				fmt.Println("main trail : ",tmpPlayer[0].Trail)*/
			}
			time.Sleep(3*time.Second)
		}

	}()
	fmt.Scanln()
}
