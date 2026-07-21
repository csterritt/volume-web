## Volume Web

This is the web server for the [CYD volume controller](https://github.com/csterritt/CYD-Volume-Control).

It does two things:

1. It serves a simple web interface to control the volume
2. It fetches weather data from the Open-Meteo API and returns it as JSON
- The weather data is cached, so it only hits Open-Meteo every ten minutes.

### Customization

Currently, the latitude and longitude are hardcoded to London, UK. To change this, edit the `defaultLat` and `defaultLon` variables in `main.go`. You'll probably want to change this to your locality.

### License

It is released under the Mozilla Public License 2.0. See `LICENSE.txt` for details.
