const sqlite3 = require('sqlite3').verbose();
const Sentry = require('@sentry/node'); // v5.15.5

// const MODIFIED_DSN_FORWARD = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:3001/2'
// const MODIFIED_DSN_SAVE = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:3001/3'

const DSN = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:9000/4'

// TODO put Node module and path for DB you'll write to
var db = new sqlite3.Database('./sqlite.db', sqlite3.OPEN_READWRITE, (err) => {
    if (err) {
      return console.error(err.message);
    }
    console.log('Connected to the in-memory SQlite database.');
});

Sentry.init({ 
    dsn: DSN,
    beforeSend: function (event, hint) {
        // TODO You should check if device is online or offline somehow
        var online = false
        if (online) {
            // send to Sentry.io
            return event
        } else {
            db.run("INSERT INTO events VALUES (?,?,?,?,?)", [8,"node","example", Buffer.from(JSON.stringify(event)),{}]);
            // do not attempt sending to Sentry.io
            return null
        }
    }
});

try {
    // This is an example of your app's js/Node having an error
    // This will get written to offline db, see beforeSend function above ^
    throw new Error('test1018');
} catch (e) {
    Sentry.captureException(e)
}
