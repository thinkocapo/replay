const sqlite3 = require('sqlite3').verbose();
const Sentry = require('@sentry/node');

// Put your DSN key here
const DSN = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:9000/4'

var eventSavedOffline;

Sentry.init({ 
    dsn: DSN,
    beforeSend: function (event,hint) {
        // Normally you return the 'event' here, but in our case we're returning the event that we loaded from sqlite (i.e. it was saved offline)
        return eventSavedOffline
    }
});

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
      
      // the 5th event in my database is a Node Event. All the rest are Python Events, at the moment.
      let row = rows[4]
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
