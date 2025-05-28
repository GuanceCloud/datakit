snmpwalk -v3 -l authPriv -u snmpv3user1 -a SHA -A "authPassAgent1" -x AES -X "privPassAgent1" udp:127.0.0.1:1161 .1.3.6.1.2.1.1 ; exit 0

#-n "myView" 
#-E "80001f8880c432727ab57b316800000000"
