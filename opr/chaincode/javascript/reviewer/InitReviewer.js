'use strict';

const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const YAML = require('yamljs');

async function main() {
    const argv = process.argv;
    const user = argv[2];
    const organization = argv[3];
    const paperKey = argv[4];

    const paperID = "".concat('\u0000', 'OI', '\u0000', organization, '\u0000', paperKey, '\u0000');

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
        const contract = network.getContract('reviewer');

        // Submit the specified transaction.
        await contract.submitTransaction('InitReviewer', paperID);
        console.log('Transaction has been submitted');

        // Disconnect from the gateway.
        await gateway.disconnect();

    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

main();
