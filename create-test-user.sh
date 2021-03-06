#!/bin/sh

XTOKEN=$1
if [ -z $XTOKEN ]; then
  echo "missing api key parameter"
  exit
fi

DATABUNKER="http://localhost:3000"

RESULT=`curl -s $DATABUNKER/v1/pactivity/share-data-with-sms-provider -XPOST \
   -H "X-Bunker-Token: $XTOKEN"  -H "Content-Type: application/json" \
   -d '{"title":"send sms","script":"<script>alert(1);</script>","fulldesc":"full desc","applicableto":"empty"}'`
echo "Create processing activity entity: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/lbasis/send-sms -XPOST \
   -H "X-Bunker-Token: $XTOKEN"  -H "Content-Type: application/json" \
   -d '{"module":"signup","fulldesc":"full","shortdesc":"short","requiredmsg":"required","usercontrol":false,"requiredflag":true}'`
echo "Create legal basis entity: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/lbasis/send-sms-on-login -XPOST \
   -H "X-Bunker-Token: $XTOKEN"  -H "Content-Type: application/json" \
   -d '{"module":"signup","fulldesc":"full","shortdesc":"short","requiredmsg":"required","usercontrol":false,"requiredflag":true}'`
echo "Create legal basis entity 2: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/lbasis/send-email-on-login -XPOST \
   -H "X-Bunker-Token: $XTOKEN"  -H "Content-Type: application/json" \
   -d '{"module":"signup","fulldesc":"full","shortdesc":"short","requiredmsg":"required","usercontrol":false,"requiredflag":true}'`
echo "Create legal basis entity 3: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/pactivity/share-data-with-sms-provider/blah -XPOST \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Tryingto link fake legal basis to processing activity: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/pactivity/share-data-with-sms-provider/send-sms-on-login -XPOST \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Linking existing legal basis 2 to processing activity: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/pactivity/share-data-with-sms-provider/send-sms-on-login -XPOST \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Linking again existing legal basis 2 to processing activity: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/pactivity/share-data-with-sms-provider/send-email-on-login -XPOST \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Linking existing legal basis 3 to processing activity: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/pactivity/share-data-with-sms-provider/send-sms -XPOST \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Linking existing legal basis to processing activity: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/pactivity/share-data-with-sms-provider/send-sms -XDELETE \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Unlinking legal basis to processing activity: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/pactivity -H "X-Bunker-Token: $XTOKEN"`
echo "Get a list of processing activities: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/lbasis -H "X-Bunker-Token: $XTOKEN"`
echo "Get a list of legal basis objects: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/lbasis/send-sms -XDELETE -H "X-Bunker-Token: $XTOKEN"`
echo "Deleting legal basis object: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/agreement/send-sms-on-login/email/test@paranoidguy.com -XPOST \
   -H "X-Bunker-Token: $XTOKEN"  -H "Content-Type: application/json" \
   -d '{"lawfulbasis":"contract"}'`
echo "Giving consent for legal basis obj 2: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/agreement/send-sms-on-login -XDELETE -H "X-Bunker-Token: $XTOKEN"`
echo "Revoking legal basis object 2: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/agreement/contract-approval/email/test@paranoidguy.com -XPOST \
   -H "X-Bunker-Token: $XTOKEN"  -H "Content-Type: application/json" \
   -d '{"lawfulbasis":"contract"}'`
echo "Giving consent for fake legal basis: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/agreement/send-sms-on-login/email/test@paranoidguy.com -XDELETE -H "X-Bunker-Token: $XTOKEN"`
echo "Withdraw legal basis 2 consent: $RESULT"

echo "Creating user."
RESULT=`curl -s $DATABUNKER/v1/user \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"fname":"Test","lname":"Account","email":"test@paranoidguy.com","phone":"4444","passportid":"123456789","status":"prospect"}'`
STATUS=`echo $RESULT | jq ".status" | tr -d '"'`
if [ "$STATUS" = "error" ]; then
  echo "Error to create user, trying to lookup by email. Result: $RESULT"
  RESULT=`curl -s $DATABUNKER/v1/user/email/test@paranoidguy.com -H "X-Bunker-Token: $XTOKEN"`
  echo "Result: $RESULT"
  STATUS=`echo $RESULT | jq ".status" | tr -d '"'`
fi
if [ "$STATUS" = "error" ]; then
  echo "Failed to fetch user by email, got: $RESULT"
  exit
fi

TOKEN=`echo $RESULT | jq ".token" | tr -d '"'`
echo "User token is $TOKEN"

RESULT=`curl -s $DATABUNKER/v1/userapp/token/$TOKEN/shipping \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"country":"UK","city":"London","address":"221B Baker Street","postcode":"12345","status":"new"}'`
echo "User shipping record created, status $RESULT"

RESULT=`curl -s $DATABUNKER/v1/userapp/token/$TOKEN/shipping -XPUT \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"status":"delivered"}'`
echo "User shipping record updated, status $RESULT"

RESULT=`curl -s $DATABUNKER/v1/userapp/token/$TOKEN/shipping \
  -H "X-Bunker-Token: $XTOKEN"`
echo "User shipping record ready, status $RESULT"

RESULT=`curl -s $DATABUNKER/v1/sharedrecord/token/$TOKEN \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"app":"shipping","fields":"address"}'`
echo "Shared record created, status $RESULT"
RECORD=`echo $RESULT | jq ".record" | tr -d '"'`
echo $RECORD

RESULT=`curl -s $DATABUNKER/v1/get/$RECORD`
echo "Get shared record (no password/access token): $RESULT"

RESULT=`curl -s $DATABUNKER/v1/userapp/token/$TOKEN \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json"`
echo "View list of all user apps $RESULT"

RESULT=`curl -s $DATABUNKER/v1/userapps \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json"`
echo "View list of all apps $RESULT"

RESULT=`curl -s $DATABUNKER/v1/agreement/send-sms/token/$TOKEN -XPOST \
   -H "X-Bunker-Token: $XTOKEN" -d "expiration=30s"`
echo "Enable consent send-sms for user by token: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/agreement/send-sms-on-login/email/test@paranoidguy.com -XPOST \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Enable consent send-sms for user by email: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/agreement/send-sms-on-login/phone/4444 -XDELETE \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Withdraw consent send-sms for user by phone: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/agreement/send-sms/token/$TOKEN \
   -H "X-Bunker-Token: $XTOKEN"`
echo "View this specific consent only: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/agreement/token/$TOKEN \
   -H "X-Bunker-Token: $XTOKEN"`
echo "View all user consents: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/agreement/send-sms \
   -H "X-Bunker-Token: $XTOKEN"`
echo "View all users with send-sms consent on: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/session/token/$TOKEN -XPOST \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
   -d '{"clientip":"1.1.1.1","x-forwarded-for":"2.2.2.2"}'`
echo "Create session 1: $RESULT"

SESSION=`echo $RESULT | jq ".session" | tr -d '"'`
echo $SESSION

RESULT=`curl -s $DATABUNKER/v1/session/email/test@paranoidguy.com -XPOST \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
   -d '{"clientip":"1.1.1.1","x-forwarded-for":"2.2.2.2","info":"email"}'`
echo "Create session 2: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/session/session/$SESSION \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json"`
echo "Get session 1: $RESULT"

RESULT=`curl -s $DATABUNKER/v1/session/phone/4444 \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json"`
echo "Get sessions: $RESULT"
