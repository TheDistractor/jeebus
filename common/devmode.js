// Generated by CoffeeScript 1.7.1
(function() {
  var execFile, fatal, gin, installGin, ready, runGin, spawn, _ref,
    __slice = [].slice;

  _ref = require('child_process'), execFile = _ref.execFile, spawn = _ref.spawn;

  fatal = function() {
    var args, s;
    s = arguments[0], args = 2 <= arguments.length ? __slice.call(arguments, 1) : [];
    console.error('\n[node] fatal error:', s);
    if (args.length) {
      console.error.apply(console, args);
    }
    return process.exit(1);
  };

  runGin = function(done) {
    var p;
    p = spawn('gin', [], {
      stdio: 'pipe'
    });
    p.on('close', function(code) {
      if (code > 0) {
        return fatal('unexpected termination of "gin", code: ' + code);
      }
    });
    p.stdout.on('data', function(data) {
      var s;
      s = data.toString();
      if (data.length > 0) {
        process.stdout.write(s);
      }
      if (/listening on port/.test(s)) {
        return ready();
      }
    });
    p.stderr.on('data', function(data) {
      var s;
      s = data.toString();
      if (!/execvp\(\)/.test(s)) {
        return process.stderr.write(s);
      }
    });
    return p;
  };

  installGin = function() {
    console.log('"gin" tool not found, installing...');
    return execFile('go', ['get', 'github.com/codegangsta/gin'], function(err, sout, serr) {
      var gin;
      if (err) {
        if (err.code === 'ENOENT') {
          fatal('"go" not found - please install it first, see http://golang.org/');
        }
        fatal('install of "gin" failed', serr);
      }
      gin = runGin();
      return gin.on('error', function(err) {
        return fatal('still cannot launch "gin" - is $GOPATH/bin in your $PATH?');
      });
    });
  };

  ready = function() {
    return console.log('[node] watching for file changes (NOT YET)');
  };

  gin = runGin();

  gin.on('error', function(err) {
    if (err.code !== 'ENOENT') {
      fatal('cannot launch "gin"', err);
    }
    return installGin();
  });

}).call(this);

//# sourceMappingURL=devmode.map
