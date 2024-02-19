# A Template for Blockchain-Based Open Peer Review Systems Using Hyperledger Fabric

We present a template for creating a configurable open peer review system as an effective tool to facilitate an open peer review process.

This system is configured on a blockchain network using Hyperledger Fabric. For convenience, we provide a prototype webpage. All activities on the webpage, where open peer review can be conducted, are stored in blocks and displayed to the subject according to the acceptance model.

## Requirements
  - Docer >= 25.0.3
  - Node.js >= 12.22.9
  - npm >= 8.5.1

```bash
#Note: This prototype template has some permission and path issues. We hope you to run it as the root user from the root directory.

# Clone this repository.
clone https://github.com/knu-plml/boprs.git /root/fabric

# Download and move the Docker image.
# The image can be downloaded from https://drive.google.com/file/d/1j2Lxy8hGwuQtmAmnh6cojll7hR3O2Tiv/view?usp=sharing.
mv default-fabric-image.tar /root/fabric/script

# Configure a blockchain network and install chaincodes using our proposed template.
cd /root/fabric/script
./makeTestNetwork.sh
./installAllCC.sh

# Install packages for web pages and chaincodes.
cd /root/fabric/opr
npm install

# Register example users in the blockchain network for the open peer review.
cd /root/fabric/opr/chaincode/javascript
./register.sh

# Running a web service.
cd /root/fabric/opr/javascript
node app.js
```

The open peer review service is hosted at localhost:3000
