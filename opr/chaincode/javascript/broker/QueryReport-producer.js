'use strict';

const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const YAML = require('yamljs');

const Kafka = require('node-rdkafka');

async function main() {
    const argv = process.argv;
    const user = 'editor@ezmeta.co.kr';
    const topicName = 'QueryReport';
    var broker = argv[7];
    const organization = 'JISTAP';
    var paperKey = argv[2];
    var round = argv[3];
    var locations = argv[4];
    var comments = argv[5];
    var decision = argv[6];
    var contractID;
    var paperID;

    if(!broker) {
        broker = 'localhost:9092';
    }
    contractID = "".concat('\u0000', '\u0000', organization, '\u0000', paperKey, '\u0000', round, '\u0000');
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
        var contract = network.getContract('report');

        // Submit the specified transaction.
//        await contract.submitTransaction('AddReport', contractID, locations, comments, decision);
//        console.log('Transaction has been submitted');

        const report = await contract.evaluateTransaction('QueryReport', contractID);
        console.log(`Transaction has been evaluated`);

        // Get the contract from the network.
        contract = network.getContract('comment');

        const commentList = await contract.evaluateTransaction('QueryCommentWithContractKey', contractID);
        console.log(`Transaction has been evaluated`);

        // Disconnect from the gateway.
        await gateway.disconnect();


        // Add PaperKey to report json data
        var report_json = JSON.parse(report);
        var commentList_json = JSON.parse(commentList);

        for(var i in commentList_json) {
            commentList_json[i].ID = commentList_json[i].Key;
            commentList_json[i].Key = undefined;
            commentList_json[i].ContractKey = undefined;
        }


        var _report_json = {};
        _report_json.ID = report_json.Key;
        _report_json.ContractID = report_json.ContractKey;
        _report_json.Decision = report_json.Decision;
        _report_json.PaperID = paperKey;
        _report_json.OverallComment = {};
        _report_json.OverallComment.ID = report_json.OverallComment.Key;
        _report_json.OverallComment.ContractID = report_json.OverallComment.ContractKey;
        _report_json.OverallComment.ReviewerID = report_json.OverallComment.ReviewerID;
        _report_json.OverallComment.Location = report_json.OverallComment.Location;
        _report_json.OverallComment.Comment = report_json.OverallComment.Comment;
        _report_json.OverallComment.ReviewerComment = commentList_json;

        // Our producer with its Kafka brokers
        // This call returns a new writable stream to our topic 'topic-name'
        var stream = Kafka.Producer.createWriteStream({
            'metadata.broker.list': broker
        }, {}, {
            topic: topicName
        });

        // Writes a message to the stream
        var queuedSuccess = stream.write(Buffer.from(JSON.stringify(_report_json)));

        if (queuedSuccess) {
            console.log(`queued ${paperID}[${round}] report to ${broker}.`);
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
