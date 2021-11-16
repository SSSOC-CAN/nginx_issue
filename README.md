# Ngnix Issue
This is a minimum, reproducible example for an issue I'm encoutering when deploying my gunicorn app to nginx.

# Installation (Linux)

## Webserver

in the `/webserver` directory, create a Python virtual environment `python3 -m venv venv` and activate it `source venv/bin/activate` then install all requirements from the `requirements.txt` file `pip install -r requirements.txt`

Make sure nginx is installed on your machine. Copy the `webserver_nginx_conf` file to `/etc/nginx/sites-available` directory. Then establish the symbolic link `sudo ln -s /etc/nginx/sites-available/myproject /etc/nginx/sites-enabled`. Perform the test command `sudo nginx -t` and then `sudo systemctl restart nginx`

# Running without Nginx

Webserver:
```
(venv) /webserver$ gunicorn --worker-class gevent --workers 1 --timeout=160 --bind 0.0.0.0:5000 webserver:app
```

Go-Server:
```
/goserver$ go run main.go
```

When visiting `http://localhost:5000` you should notice in the browser console that data is coming in and is being properly parsed.

# Running with Nginx

Make sure Nginx is running
Webserver:
```
(venv) /webserver$ gunicorn --worker-class gevent --workers 1 --timeout=160 --bind 0.0.0.0:5000 webserver:app
```

Go-Server:
```
/goserver$ go run main.go
```

Now when checking browser console, some responses should get truncated.