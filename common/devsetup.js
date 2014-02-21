// Generated by CoffeeScript 1.7.1
(function() {
  var JEEBUS_ROOT, execFile, fatal, fs, goDir, installJeeBus, jbDir, path, readline, rl, _ref,
    __slice = [].slice;

  console.log('\nThis script sets up a fresh application area based on JeeBus.\nYou need to supply the name of a new directory to initialise.\nIt will be prepared with a minimal set of files and settings.\n');

  fs = require('fs');

  execFile = require('child_process').execFile;

  readline = require('readline');

  path = require('path');

  JEEBUS_ROOT = (_ref = process.env.JEEBUS_ROOT) != null ? _ref : 'github.com/jcw/jeebus';

  fatal = function() {
    var args, s;
    s = arguments[0], args = 2 <= arguments.length ? __slice.call(arguments, 1) : [];
    console.error('\n[node] fatal error:', s);
    if (args.length) {
      console.error.apply(console, args);
    }
    return process.exit(1);
  };

  if (!process.env.GOPATH) {
    fatal('GOPATH undefined, please make sure "go" has been installed');
  }

  goDir = process.env.GOPATH.split(':')[0];

  jbDir = "" + goDir + "/src/" + JEEBUS_ROOT;

  installJeeBus = function(done) {
    if (fs.existsSync(jbDir)) {
      return done();
    } else {
      console.log("Fetching jeebus from https://" + JEEBUS_ROOT);
      return execFile('go', ['get', JEEBUS_ROOT], function(err, sout, serr) {
        if ((err != null ? err.code : void 0) === 'ENOENT') {
          fatal('"go" not found - please install it first, see http://golang.org/');
        }
        if (err) {
          fatal('fetching failed', serr);
        }
        if (!fs.existsSync(jbDir)) {
          fatal('still cannot find jeebus');
        }
        return done();
      });
    }
  };

  rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
  });

  rl.question('Directory name? ', function(appDir) {
    var name, title;
    rl.close();
    name = path.basename(appDir);
    title = name.charAt(0).toUpperCase() + name.slice(1).toLowerCase();
    if (fs.existsSync(appDir)) {
      fatal('please enter the name of a nonexistent directory to initialise');
    }
    return installJeeBus(function() {
      fs.mkdirSync(appDir);
      fs.mkdirSync("" + appDir + "/app");
      fs.writeFileSync("" + appDir + "/index.js", "require('" + jbDir + "/common/devmode');\n");
      fs.writeFileSync("" + appDir + "/settings.txt", "COMMON_DIR = \"" + jbDir + "/common\"\n");
      fs.writeFileSync("" + appDir + "/package.json", "{\n  \"name\": \"" + name + "\",\n  \"description\": \"This is the new " + title + " application.\",\n  \"version\": \"0.0.1\",\n  \"main\": \"index.js\"\n}\n");
      fs.writeFileSync("" + appDir + "/main.go", "package main\n\nimport (\n    \"log\"\n    \"" + JEEBUS_ROOT + "\"\n)\n\nconst Version = \"0.0.1\"\n\nfunc init() {\n    log.SetFlags(log.Ltime) // only display HH:MM:SS time in log entries\n}\n\nfunc main() {\n    println(\"\\n" + title + "\", Version, \"/ JeeBus\", jeebus.Version)\n    jeebus.Run()\n}\n");
      fs.writeFileSync("" + appDir + "/app/index.html", "<!DOCTYPE html>\n<html>\n<head>\n  <meta http-equiv='Content-type' content='text/html; charset=utf-8'>\n  <title>" + title + "</title>\n</head>\n<body>\n  <h1>Welcome to " + title + " !</h1>\n</body>\n");
      return console.log("\n" + title + " has been created. To start it up, enter this command:\n    cd " + appDir + " && node .\nThen open the web page at http://localhost:3000/ - that's it!\n");
    });
  });

}).call(this);

//# sourceMappingURL=devsetup.map