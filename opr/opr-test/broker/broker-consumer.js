'use strict';
//var nodemailer = require('nodemailer')
//
//function sendEmail(sender, receiver, paper) {
//  var transporter = nodemailer.createTransport({
//    host: "smtp.gmail.com",
//    secure: true,
//    auth: {
//      user: 'opr.testmail@gmail.com',
//      pass: 'pl@8441308307'
//    }
//  })
//
//  var subject = '[BOPR] You was selected a reviewer.'
//  var content = "You was selected a reviewer on the paper '" + paper + "'.\n\nVisit the website and accept to be a r    eviewer. \n\nhttp://39.115.145.90:3000"
//
//  var mailOptions = {
//    from: sender + ' <opr.testmail@gmail.com>',
//    to: receiver,
//    subject: subject,
//    text: content
//  }
//
//  transporter.sendMail(mailOptions, function(error, info){
//    if (error) {
//      console.log(error);
//    }
//    else {
//      console.log('Email sent: ' + info.response);
//    }
//  })
//}

const Transform = require('stream').Transform;
const Kafka = require('node-rdkafka');

const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const YAML = require('yamljs');

async function main() {
    const argv = process.argv;
    const user = 'broker';
    var topic = ['AddPaper', 'RegisterCandidate'];

    var broker = argv[2];

    if(!broker) {
        broker = 'localhost:9092';
    }

    if (broker == 'localhost:9092') {
        topic = ['AddPaper', 'RegisterCandidate', 'SelfRecList', 'QueryReport'];
    }

    var consumer = new Kafka.KafkaConsumer({
        //'debug': 'all',
        'metadata.broker.list': broker,
        'group.id': 'boprs',
        'enable.auto.commit': true,
        'receive.message.max.bytes': 104857600,
        'fetch.message.max.bytes': 104857600
    });

    //logging debug messages, if debug is enabled
    consumer.on('event.log', function(log) {
        console.log(log);
    });

    //logging all errors
    consumer.on('event.error', function(err) {
        console.error('Error from consumer');
        console.error(err);
    });

    consumer.on('ready', function(arg) {
        console.log('consumer ready.' + JSON.stringify(arg));

        consumer.subscribe(topic);
        console.log(`subcribe ${broker} ${topic}.`);
        //start consuming messages, read 10 message every 1000 miliseconds.
        setInterval(function() {
            consumer.consume(10);
        }, 3000);
    });


    consumer.on('data', function(m) {
        try{
            console.log(m.topic);
            var json = JSON.parse(m.value.toString());
            json.PaperFile = "PaperFile";
            console.log(json);
            submit(user, m);
        } catch(e) {
            console.log(`"${m.value.toString()}" dose have the JSON format.`);
        }
    });

    consumer.on('disconnected', function(arg) {
        console.log('consumer disconnected. ' + JSON.stringify(arg));
    });

    //starting the consumer
    consumer.connect();
}

async function submit(user, message) {
    try {
        const ccpPath = path.resolve('/', 'root', 'fabric', 'script', 'connection.yaml');
        const ccp = YAML.parse(fs.readFileSync(ccpPath, 'utf8'));

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        const identity = await wallet.get(user);
        if (!identity) {
            console.log(`An identity for the user ${user} does not exist in the wallet`);
            console.log(`Run the registerUser.js application before retrying`);
            return;
        }

        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        await gateway.connect(ccp, { wallet, identity: user, discovery: { enabled: false, asLocalhost: true } });

        // Get the network (channel) our contract is deployed to.
        const network = await gateway.getNetwork('boprs');

        if(message.topic == 'AddPaper') {
            const paper = JSON.parse(message.value.toString());
            // load the network configuration
            // Get the contract from the network.
            const contract = network.getContract('paper');

            // Submit the specified transaction.
            await contract.submitTransaction('AddPaperWithEditorID', paper.Organization, paper.AuthorID, paper.PaperID, paper.Title, paper.PaperFile, paper.AuthorEmail, paper.Abstract, paper.EditorID);
            console.log(`${paper.Organization} ${paper.PaperID} paper Transaction has been submitted`);


        } else if(message.topic == 'RegisterCandidate') {
            const candidateInfo = JSON.parse(message.value.toString());
            const paperID = "".concat('\u0000', 'OI', '\u0000', candidateInfo.Organization, '\u0000', candidateInfo.PaperID, '\u0000');
            const CandidateList= candidateInfo.CandidateList;

            // Get the contract from the network.
            const contract = network.getContract('reviewer');

            // Submit the specified transaction.
            for(var i in CandidateList) {
                try {
                    //sendEmail(sender, CandidateList[i].email, paperName)
                    await contract.submitTransaction('RegisterReviewer', paperID, CandidateList[i].ORCID, CandidateList[i].email);
                    console.log(`Transaction has been submitted. ${CandidateList[i].ORCID} was registered reviewer of ${paperID} paper.`);
                } catch (error) {
                    console.error(`Failed to submit transaction${CandidateList[i].ORCID}: ${error}`);
                }
            }
//        }
        } else if(message.topic == 'SelfRecList') {

            const candidateList = message.value.toString();
    
            // Our producer with its Kafka brokers
            // This call returns a new writable stream to our topic 'topic-name'
            var stream = Kafka.Producer.createWriteStream({
                'metadata.broker.list': 'localhost:9092'
            }, {}, {
                topic: 'RegisterCandidate'
            });
    
            // Writes a message to the stream
            var queuedSuccess = stream.write(Buffer.from(candidateList));
    
            if (queuedSuccess) {
                console.log(`queued ${candidateList}.`);
            } else {
                // Note that this only tells us if the stream's queue is full,
                // it does NOT tell us if the message got to Kafka!  See below...
                console.log('Too many messages in our queue already');
            }
    
            stream.close();
        }

        // Disconnect from the gateway.
        await gateway.disconnect();
    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
    }
}

main();
