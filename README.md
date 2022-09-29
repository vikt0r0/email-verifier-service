Email verifier service
======================

Go package which uses a modified version of [Aftership's Email Library](https://github.com/AfterShip/email-verifier).

Copy the constants file with `cp apiserver/constants.go.template apiserver/constants.go` and edit `apiserver/constants.go` to your needs.

To build the service binary, run `go build`. To run the service, do `./email-verifier-service start --addr="$(curl -s ifconfig.me):3002"`. In this case, you will bind the service to the public IP if applicable. If you are behind a router with NAT, you should enable port forwarding and bind to your network interface IP, or use a tunnel.

To verify an email, do a GET request on the binding IP on the designated port, e.g. `curl -s -X GET 'http://localhost:3002?email=email@totest.com'`. The response is a JSON string which looks something like this:

	{
	  "Result": {
	    "email": "",
	    "reachable": "unknown",
	    "syntax": {
	      "username": "",
	      "domain": "",
	      "valid": false
	    },
	    "smtp": null,
	    "gravatar": null,
	    "suggestion": "",
	    "disposable": false,
	    "role_account": false,
	    "free": false,
	    "has_mx_records": false
	  },
	  "Error": true,
	  "ErrorMessage": "no or empty email GET parameter specified"
	}

If "Error" is true, a network error occured or no email get parameter was specified. Otherwise there is a result, where syntax pertains to syntax validation of the email (i.e. does it have the correct form), smtp contains a JSON object that contains information about whether the domain associated with the email has valid email servers and whether it could be contacted. The most crucial field, however is the reachable field which has the value "true", "false" or "unknown". In case of unknown perhaps it makes sense to delegate the request to a more intelligent service such as mailgun. But this service can hopefully limit the number of requests.


Also ensure that port 25 outgoing is open, otherwise the script will not be able to perform SMTP checks.

Be careful that your email setup is correct, so you do not get your IP or domain name banned. A good place to check it is here [https://multirbl.valli.org/lookup/142.250.74.110.html](https://multirbl.valli.org/lookup/142.250.74.110.html) (but edit the string to your IP instead).

systemd daemon
--------------
If you plan to run the service persistently, you can consider adding a service to systemd (if you are running Ubuntu), do `touch /etc/systemd/system/email-service.service` and add something like:

	[Unit]
	Description=Email Checker Service Daemon
	After=network.target
	StartLimitIntervalSec=0
	[Service]
	Type=simple
	Restart=always
	RestartSec=1
	User=ubuntu
	#ExecStart=/opt/email-verifier-service/email-verifier-service start --addr="$(curl -s ifconfig.me):3002"
	ExecStart=/opt/email-verifier-service/email-verifier-service start --addr="127.0.0.1:3002"

	[Install]
	WantedBy=multi-user.target

but modify it appropriately. Then do `sudo systemd enable email-service` and `sudo systemd enable email-service`.
