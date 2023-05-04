#! /bin/sh
# run this to simulate load tests
# 30309 is the zip code of the Margaret Mitchell House
mongosh 'mongodb://localhost:30309/' --eval 'db.shutdownServer()'
rm -rf dbase
mkdir dbase

mongod --dbpath dbase --logpath dbase/mongod.log --port 30309 --fork
sleep 5
mongosh 'mongodb://localhost:30309/' --eval 'db.setProfilingLevel(0, 10)'

hatchet --bios 'mongodb://localhost:30309/' 10000000 &

for i in {1..3}
do
    hatchet --sim read 'mongodb://localhost:30309/' &
    hatchet --sim write 'mongodb://localhost:30309/' &
done

wait

hatchet --obfuscate dbase/mongod.log | gzip > obfuscated.log.gz
mongosh 'mongodb://localhost:30309/' --eval 'db.shutdownServer()'
