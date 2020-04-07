// ***************************************************************************************
// Use web3.js (not GoChain's CLI) to interact with the blockchain and query logged events 
// associated with the contract
// ***************************************************************************************

// Require the node web3.js library (to interact with Ethereum based networks)
var Web3 = require('web3');

// Create a web3 object pointing to the GoChain mainnet
var web3 = new Web3("https://rpc.gochain.io/");

const START_BLOCK = 11958580;
const END_BLOCK = 11923080;

// Specify the target contract's ABI
//   the ABI describes a contract's specification (I think) and is produced when compiled
var myContractABI = [
    {
      "constant": false,
      "inputs": [
        {
          "name": "hsh",
          "type": "string"
        },
        {
          "name": "go_ref",
          "type": "string"
        },
        {
          "name": "cust_ref",
          "type": "string"
        }
      ],
      "name": "postObj",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "constructor"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "name": "logHash",
          "type": "string"
        },
        {
          "indexed": true,
          "name": "logGoRef",
          "type": "string"
        },
        {
          "indexed": false,
          "name": "logCustRef",
          "type": "string"
        }
      ],
      "name": "LogVal",
      "type": "event"
    }
  ];

// Create a contract object referencing the deployed contract
var myContract = new web3.eth.Contract(myContractABI, '0x4B6895bA8495d4920F40b6b8dCB1361621808912');

// Output the events associated with the contract between two blocks (in this case a specified block and the last block)
myContract.getPastEvents("allEvents", 
{
    fromBlock: START_BLOCK,
    toBlock: 'latest'
}).then(events => console.log(events))
.catch((err) => console.error(err));