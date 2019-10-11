# sportmux
Serial port to IP multiplexer

sportmux connects to a serial port and listens to a TCP port. Any number
of clients can connect to the TCP port. Everything that is read from the
serial port is relayed to all TCP clients and anything that is written by
the TCP clients are relayed to the serial port.
