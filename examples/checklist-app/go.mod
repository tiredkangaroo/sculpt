module main

go 1.23

toolchain go1.23.0

require github.com/tiredkangaroo/sculpt v1.0.0

require github.com/lib/pq v1.10.9 // indirect

replace github.com/tiredkangaroo/sculpt => ../../

replace github.com/tiredkangaroo/sculpt/adminpanel => ../../adminpanel
