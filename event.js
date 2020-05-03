const sqlite3 = require('sqlite3').verbose();
const Sentry = require('@sentry/node'); // v5.15.5


const ORIGINAL_DSN = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:9000/4'

const MODIFIED_DSN_FORWARD = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:3001/2'
const MODIFIED_DSN_SAVE = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:3001/3'

var db = new sqlite3.Database('./sqlite.db', sqlite3.OPEN_READWRITE, (err) => {
    if (err) {
      return console.error(err.message);
    }
    console.log('Connected to the in-memory SQlite database.');
});

Sentry.init({ 
    dsn: MODIFIED_DSN_SAVE,
    beforeSend: function (event, hint) {
        var online = false

        if (online) {
            // send to Sentry.io
            return event
        } else {
            // save to Sqlite or other offline database
            
            return null
        }
    }
});

// console.log('Sentry.ca', Sentry.captureException)
// throw new Error('test');

// new Error('this is it')
// Sentry.captureException(new Error('this is the error'))


try {
    throw new Error('test713');
} catch (e) {
    console.log('\nE\n', typeof(e))
    Sentry.captureException(e)
    // Sentry.captureException(new Error('Hello There'));
    // Sentry.captureMessage("This is The Test");

}