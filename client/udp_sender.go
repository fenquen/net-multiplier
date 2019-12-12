package client

type UdpSender struct {
	SenderBase
	//remoteAddr net.Addr
}

/*func (udpSender *UdpSender) Write(byteSlice []byte) (int, error) {
	udpConn, _ := udpSender.conn2DestSvr.(*net.UDPConn)
	return udpConn.WriteToUDP(byteSlice, udpSender.remoteAddr)
}*/
