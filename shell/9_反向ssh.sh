#!/bin/bash
## ------------- debian systemd -------
#[Unit]
#Description= reverse ssh Daemon
#After=multi-user.target
#
#[Service]
#ExecStart=/path/to/reverse_ssh.sh
#Restart=always
#
#[Install]
#WantedBy=multi-user.target
#



while true
do
	ps=$(ps -eo cmd |grep -v grep |grep '55555:localhost:22')
	if [ -z "$ps"  ]
       	then
		echo "restart ssh server"
		ssh -p 3389 -fNR 55555:localhost:22 zh@my_vps
	fi
	sleep 30
done


