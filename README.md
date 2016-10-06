[![GoDoc](https://godoc.org/github.com/qarea/redminems?status.svg)](https://godoc.org/github.com/qarea/redminems)
# Redmine-adapter microservice

Project uses [narada framework](https://github.com/powerman/Narada)  
Based on [go-socklog](https://github.com/powerman/narada-base/tree/go-socklog) template  
Exteneded with narada [go-plugin](https://github.com/powerman/narada-plugin-go-service)

You have to install [runit](http://smarden.org/runit/) and [socklog](http://smarden.org/socklog/)
If services didn't start after `./release && ./deploy` check if runit, socklog, perl libs and narada tools are reachable for user and cron.  

