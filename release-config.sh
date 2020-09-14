# must run this from your directory that has .git, where you isntalled sentry-demos/tracing.
day=$(date +%d)
month=$(date +%-m)

if [ "$day" -ge 0 ] && [ "$day" -le 7 ]; then
  echo ">=0<=7"
  week=1
elif [ "$day" -ge 8 ] &&  [ "$day" -le 14 ]; then
  echo ">=8<=14"
  week=2
elif [ "$day" -ge 15 ] &&  [ "$day" -le 21 ]; then
  echo ">=15<=21"
  week=3
elif [ "$day" -ge 22 ]; then
  echo ">=22"
  week=4
fi

RELEASE="$month.$week"
echo $release

# SENTRY_AUTH_TOKEN defined in shell profile
SENTRY_PROJECT1=da-react
SENTRY_PROJECT2=da-flask
SENTRY_ORG=testorg-az
PREFIX=static/js

sentry-cli releases -o $SENTRY_ORG new -p $SENTRY_PROJECT1 $RELEASE
sentry-cli releases -o $SENTRY_ORG new -p $SENTRY_PROJECT2 $RELEASE

sentry-cli releases -o $SENTRY_ORG -p $SENTRY_PROJECT1 set-commits --auto $RELEASE
sentry-cli releases -o $SENTRY_ORG -p $SENTRY_PROJECT2 set-commits --auto $RELEASE

sentry-cli releases -o $SENTRY_ORG -p $SENTRY_PROJECT1 files $RELEASE \
		upload-sourcemaps --url-prefix "~/$PREFIX" --validate react/build/$PREFIX

echo 'DONE'
echo $release
