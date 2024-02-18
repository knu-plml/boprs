'use strict';

const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const YAML = require('yamljs');

const Kafka = require('node-rdkafka');

async function main() {
    const argv = process.argv;
    const user = 'editor';
    const topicName = 'QueryReviewerReput';
    var broker = argv[5];
    const organization = 'JISTAP';
    var paperKey = argv[2];
    var reviwerIDList = argv[3];
    var rateList = argv[4];
    var paperID;

    if(!broker) {
        broker = 'localhost:9092';
    }
    paperID = "".concat('\u0000', 'OI', '\u0000', organization, '\u0000', paperKey, '\u0000');

    try {
        // load the network configuration
        const ccpPath = path.resolve(process.cwd(), 'connection.yaml');
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

        // Get the contract from the network.
        const contract = network.getContract('rating');

        // Submit the specified transaction.
        await contract.submitTransaction('EOPRateReviewer', paperID, reviwerIDList, rateList);
        console.log(`Transaction has been submitted. Rate the reviewer of ${paperID} paper.`);

        const result = await contract.evaluateTransaction('QuerySimpleRatingWithPaperKey', paperID);
        console.log(`Transaction has been evaluated`);

        // Disconnect from the gateway.
        await gateway.disconnect();

        //const rateList = JSON.parse('{\"rateList\" : ' + rateListByte.toStrint() + '}');
        rateList = JSON.parse(result.toString());

        if(rateList[0].AuthorRating == '' || rateList[0].EditorRating == '') {
            return;
        }

        // Our producer with its Kafka brokers
        // This call returns a new writable stream to our topic 'topic-name'
        var stream = Kafka.Producer.createWriteStream({
            'metadata.broker.list': broker
        }, {}, {
            topic: topicName
        });

        for(var i in rateList) {
//            const message = '{\"ORCID\":\"' + rateList[i].ORCID + '\",\"AuthorRating\":\"' + rateList[i].AuthorRating + '\",\"EditorRating\":\"' + rateList[i].EditorRating + '\"}';
            const message = JSON.stringify(rateList[i]);

            // Writes a message to the stream
            var queuedSuccess = stream.write(Buffer.from(message));

            if (queuedSuccess) {
                console.log(`queued ${rateList[i].ORCID} rating.`);
            } else {
                // Note that this only tells us if the stream's queue is full,
                // it does NOT tell us if the message got to Kafka!  See below...
                console.log('Too many messages in our queue already');
            }
        }
        stream.close();
    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

main();
