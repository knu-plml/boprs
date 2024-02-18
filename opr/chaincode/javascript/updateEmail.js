'use strict';

const { Gateway, Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const fs = require('fs');
const path = require('path');
const YAML = require('yamljs');

async function main() {
    const argv = process.argv;
    const user = argv[2];
    const email = argv[3];
    const msp = 'JISTAP';//argv[4];
    const aff = 'jistap';//argv[5];
    try {
        // load the network configuration
        const ccpPath = path.resolve('/', 'root', 'fabric', 'script', 'connection.yaml');
        const ccp = YAML.parse(fs.readFileSync(ccpPath, 'utf8'));

        // Create a new CA client for interacting with the CA.
        const caURL = ccp.certificateAuthorities.ca.url;
        const ca = new FabricCAServices(caURL);

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        const userIdentity = await wallet.get(user);
        if (!userIdentity) {
            console.log('An identity for the user ${user} already exists in the wallet');
            return;
        }

        // Check to see if we've already enrolled the admin user.
        const adminIdentity = await wallet.get('admin');
        if (!adminIdentity) {
            console.log('An identity for the admin user "admin" does not exist in the wallet');
            console.log('Run the enrollAdmin.js application before retrying');
            return;
        }

        // build a user object for authenticating with the CA
        const provider = wallet.getProviderRegistry().getProvider(adminIdentity.type);
        const adminUser = await provider.getUserContext(adminIdentity, 'admin');
        const User = await provider.getUserContext(userIdentity, user);

        // Register the user, enroll the user, and import the new identity into the wallet.
        const identityService = ca.newIdentityService();
        const response = await identityService.update(user, { 
            affiliation: aff,
            enrollmentID: user,
            attrs: [ { name: "email", value: email, ecert: true} ]
        }, adminUser);
        console.log(response.result.attrs);
        console.log(`Successfully update ${user} email`);

        const enrollment = await ca.reenroll(User, [{ name: "userType", optional: false }, { name: "identifier", optional: false }, { name: "email", optional: true }]);
        const x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: msp,
            type: 'X.509',
        };
        await wallet.put(user, x509Identity);
        console.log(`Successfully update and reenrolled admin user ${user} and imported it into the wallet`);

        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        await gateway.connect(ccp, { wallet, identity: user, discovery: { enabled: false, asLocalhost: true } });

        // Get the network (channel) our contract is deployed to.
        const network = await gateway.getNetwork('boprs');

        // Get the contract from the network.
        const contract = network.getContract('reviewer');

        // Submit the specified transaction.
        await contract.submitTransaction('UpdateEmail', user, email);
        console.log('Transaction has been submitted');

        // Disconnect from the gateway.
        await gateway.disconnect();

    } catch (error) {
        console.error(`Failed to register user ${user}: ${error}`);
        process.exit(1);
    }
}

main();
