const sqlite3 = require('sqlite3').verbose();
const Sentry = require('@sentry/node');

// TODO Put your DSN key here
const DSN = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:9000/4'

var eventSavedOffline;

Sentry.init({ 
    dsn: DSN,
    beforeSend: function (event,hint) {
        // We're re-purposing beforeSend to send our event loaded from offline database, as opposed to the event generated below by Sentry.captureException('ignore me')
        return eventSavedOffline
        // `return event` or `return null`
    }
});

// TODO put Node module and path for DB you'll write to
let db = new sqlite3.Database('./sqlite.db', sqlite3.OPEN_READWRITE, (err) => {
    if (err) {
      return console.error(err.message);
    }
    console.log('Connected to the in-memory SQlite database.');
});

db.serialize(() => {
    db.all(`SELECT * FROM events`, [], (err, rows) => {
      if (err) {
        console.error(err.message);
      }
      
      // Selecting the most recent event I wrote to the database
      let row = rows[7]
      let string = new Buffer(row.data).toString('ascii')
      eventSavedOffline = JSON.parse(string)

      Sentry.captureException('ignore me')
    });
});
  
db.close((err) => {
    if (err) {
        return console.error(err.message);
    }
    console.log('Close the database connection.');
});
