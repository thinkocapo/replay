const Sentry = require('@sentry/node');
const DSN = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:9000/4'

Sentry.init({ 
    dsn: DSN
});


// TODO - load event from db

try {
    throw new Error('ignore me');
} catch (e) {
    console.log('ignore the thrown error, and use one loaded from database')
    
    Sentry.captureException(errorSavedOffline)
}