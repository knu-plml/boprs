'use strict';

const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const YAML = require('yamljs');

const Kafka = require('node-rdkafka');

async function main() {
    const argv = process.argv;
    const user = 'broker';
    const topicName = 'AddPaper';
    var broker = argv[9];
    const organization = 'JISTAP';
    var authorID = argv[2];
    var authorEmail = argv[3];
    var paperID = argv[4];
    var paperTitle = argv[5];
    var paperAbstract = argv[6];
    var paperFile = argv[7];
    var editorID = argv[8];

    if(!broker) {
        broker = 'localhost:9092';
    }
    try {
        const paper = { Organization : organization, AuthorID : authorID, PaperID : paperID, Title : paperTitle, EditorID : editorID, PaperFile : paperFile, AuthorEmail : authorEmail, Abstract : paperAbstract };
        const message = JSON.stringify(paper);

        // Our producer with its Kafka brokers
        // This call returns a new writable stream to our topic 'topic-name'
        var stream = Kafka.Producer.createWriteStream({
            'metadata.broker.list': broker
        }, {}, {
            topic: topicName
        });

        // Writes a message to the stream
        var queuedSuccess = stream.write(Buffer.from(message));

        if (queuedSuccess) {
            console.log(`queued ${organization}-${paperID}.`);
        } else {
            // Note that this only tells us if the stream's queue is full,
            // it does NOT tell us if the message got to Kafka!  See below...
            console.log('Too many messages in our queue already');
        }

        stream.close();
    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

main();
