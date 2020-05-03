const sqlite3 = require('sqlite3').verbose();

const Sentry = require('@sentry/node');
const DSN = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:9000/4'

// Sentry.init({ 
//     dsn: DSN
// });


// TODO - load event from db
// open database in memory
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
      console.log("\nROWS LENGTH", rows.length)
    //   console.log(rows)
    });
});
  
  
  
// close the database connection
db.close((err) => {
    if (err) {
        return console.error(err.message);
    }
    console.log('Close the database connection.');
});

// try {
//     throw new Error('ignore me');
// } catch (e) {
//     console.log('ignore the thrown error, and use one loaded from database') 
//     Sentry.captureException(errorSavedOffline)
// }