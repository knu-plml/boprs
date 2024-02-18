rm -f wallet/*

node enrollAdmin.js

node registerUser.js reviewer1@kangwon.ac.kr reviewer1Pass reviewer reviewer1@kangwon.ac.kr 'reviewer1@kangwon.ac.kr'
node registerUser.js reviewer2@kangwon.ac.kr reviewer2Pass reviewer reviewer2@kangwon.ac.kr 'reviewer2@kangwon.ac.kr'
node registerUser.js reviewer3@kangwon.ac.kr reviewer3Pass reviewer reviewer3@kangwon.ac.kr 'reviewer3@kangwon.ac.kr'

node registerUser.js author@kangwon.ac.kr authorPass reviewer author@kangwon.ac.kr 'author@kangwon.ac.kr'
node registerUser.js editor@kangwon.ac.kr editorPass editor editor@kangwon.ac.kr 'editor@kangwon.ac.kr'
node registerUser.js eic@kangwon.ac.kr eicPass eic eic@kangwon.ac.kr 'eic@kangwon.ac.kr'
node registerUser.js broker brokerPass broker broker@kangwon.ac.kr 'broker@kangwon.ac.kr'


node registerUser.js reviewer1@ezmeta.co.kr reviewer1Pass reviewer 0000-0002-8092-9949 'reviewer1@ezmeta.co.kr'
node registerUser.js reviewer2@ezmeta.co.kr reviewer2Pass reviewer 0000-0002-1821-6438 'reviewer2@ezmeta.co.kr'
node registerUser.js reviewer3@ezmeta.co.kr reviewer3Pass reviewer 0000-0001-9624-2423 'reviewer3@ezmeta.co.kr'
node registerUser.js reviewer4@ezmeta.co.kr reviewer4Pass reviewer 0000-0002-2674-9509 'reviewer4@ezmeta.co.kr'
node registerUser.js reviewer5@ezmeta.co.kr reviewer5Pass reviewer 0000-0001-8854-3154 'reviewer5@ezmeta.co.kr'
node registerUser.js author@ezmeta.co.kr authorPass reviewer author@ezmeta.co.kr 'author@ezmeta.co.kr'
node registerUser.js editor@ezmeta.co.kr editorPass editor editor@ezmeta.co.kr 'editor@ezmeta.co.kr'
node registerUser.js eic@ezmeta.co.kr eicPass eic eic@ezmeta.co.kr 'eic@ezmeta.co.kr'

node acceptanceModel/AddAcceptanceModel.js eic@kangwon.ac.kr 1 4 1 4 1 1 1
