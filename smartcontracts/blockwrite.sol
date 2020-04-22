pragma solidity ^0.5.7;

// BlockWriteSimple logs rpachain hash data to the GoChain blockchain memorializing data
contract BlockWriteSimple {
   address owner;

   // executed with the contract is deployed
   constructor() public {
       owner = msg.sender;
   }

    // postObj is a contract function that fires the event logging
    function postObj(string memory hsh, string memory go_ref, string memory cust_ref) public {
        require(owner == msg.sender, "calling address must be the contract owner/EOA");
        require(bytes(hsh).length != 0, "error: empty hash value");
        require(bytes(go_ref).length != 0, "error: empty rpachain reference value");

        emit LogVal(hsh, go_ref, cust_ref);

        return;
    }

   // LogVal defines the event to be logged
   event LogVal(
       string indexed logHash,
       string indexed logGoRef,
       string logCustRef
   );
}

