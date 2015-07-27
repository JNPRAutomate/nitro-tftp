nitro-tftp
==========

Super "fast" TFTP server

A golang based TFTP server that makes it super simple to use across all major platforms.

Usage
-----

```

```

Goals
-----

-	Maximize performance for sending data over TFTP
-	A multi-platform server: Mac (i386/amd64), Windows (i386/amd64), Linux (i386/amd64/arm), FreeBSD(i386/amd64)
-	Consistant user experience across all platforms
-	Allows for running in a daemon mode or on demand
-	Flexible configuration options availabe as a config file or with command line switches

TFTP Performance challenges
===========================

Block Size
----------

The default size of a TFTP packet is 512B blocks.

Packet loss
-----------

Excessive packet loss is a leading factor in slow TFTP transfers.

Packet overrun
--------------

Other issues that can lead to slowdown are disk I/O issues on both ends of the connection.

QoS/CoS Options
---------------
