Set up a basic web server that uses the Fiber web framework. See @https://gofiber.io for details.

The server should have the following endpoints:
- POST /api/v1/volume-up
- POST /api/v1/volume-down
- POST /api/v1/mute

The response should be a JSON object with the following structure:
{
    "success": true
}

On failure, the response should be a JSON object with the following structure:
{
    "success": false,
    "error": "error message"
}

The server should use a file at /Users/chris/tmp/volume.json to store the current volume and
mute state.

If the file does not exist, the server should run the following shell command:

osascript -e 'set ovol to output volume of (get volume settings)'

To retrieve the current volume, store that, and set mute to false.

When the server starts, if the file exists, it should read the file and set the volume and
mute state accordingly.

It sets the volume by doing a shell command to the following format:

osascript -e "set volume output volume 25"

It mutes the volume by doing a shell command to the following format:

osascript -e "set volume output muted 1"

And unmutes the volume by doing a shell command to the following format:

osascript -e "set volume output muted 0"

So when a "change" request is received, the server should read the file, update the volume,
and write the file back. Then it should do the shell command to set the volume.
