'use strict';

const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const YAML = require('yamljs');

exports.main = async function main(user, OI, OR, OP, ON, OM, OC, OPl) {
    //const argv = process.argv;
    //const user = argv[2];
    //const OI = argv[3];
    //const OR = argv[4];
    //const OP = argv[5];
    //const ON = argv[6];
    //const OM = argv[7];
    //const OC = argv[8];
    //const OPl = argv[9];

    try {
        // load the network configuration
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

        // Get the contract from the network.
        const contract = network.getContract('acceptanceModel');

        // Submit the specified transaction.
        await contract.submitTransaction('AddAcceptanceModel', OI, OR, OP, ON, OM, OC, OPl);
        console.log('Transaction has been submitted');

        // Disconnect from the gateway.
        await gateway.disconnect();

    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        throw new Error('Failed')
        //process.exit(1);
    }
}

//main();
