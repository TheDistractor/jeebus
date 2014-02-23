# JeeBus example

This area illustrates how to set up a Go application based on JeeBus.

For development, you need to install `node.js` and run `npm install`.  
All `.coffee`, `.jade`, and `.styl` files in the `app`, `base`, and `common`  
directories will automatically be recompiled whenever they change.

See the `settings.txt` files for the actual location of these directories.

To launch in development mode, type: "**`node .`**" (including the dot)

For production mode, `node.js` and `npm` are not used.  
In this case, compile and launch the app with: "**`go run main.go`**"

The [homepage][H], [discussion forum][F], and [issue tracker][I] are at JeeLabs.

[H]: http://redmine.jeelabs.org/projects/jeebus/wiki
[F]: http://jeelabs.net/projects/cafe/boards/9
[I]: http://jeelabs.net/projects/development/issues
