## Set up the volume web server to retrieve weather data every 15 minutes

The web server will maintain a cache of the weather data.
On startup, it will create a goroutine to hit the open-meteo API to retrieve the weather data,
every 10 minutes. When that succeeds, it will update the cache, using a mutex to protect the cache.

Then create a new endpoint at /weather that will return the cached weather data, along with a timestamp.
The timestamp should be the current time.
