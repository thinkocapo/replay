// npm install @sentry/node@5.14.1
const Sentry = require('@sentry/node');


const ORIGINAL_DSN = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:9000/4'

const MODIFIED_DSN_FORWARD = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:3001/2'
const MODIFIED_DSN_SAVE = 'http://211dbbd50f41437a83316cdd4bec7513@localhost:3001/3'

Sentry.init({ 
    dsn: MODIFIED_DSN_SAVE
});

// console.log('Sentry.ca', Sentry.captureException)
// throw new Error('test');

// new Error('this is it')
// Sentry.captureException(new Error('this is the error'))


try {
    throw new Error('test525');
} catch (e) {
    console.log('\nE\n', typeof(e))
    Sentry.captureException(e)
    // Sentry.captureException(new Error('Hello There'));
    // Sentry.captureMessage("This is The Test");

}