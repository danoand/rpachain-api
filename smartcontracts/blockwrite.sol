pragma solidity ^0.5.7;

// BlockWriteSimple logs rpachain hash data to the GoChain blockchain memorializing data
contract BlockWriteSimple {
   address owner;

   // executed with the contract is deployed
   constructor() public {
       owner = msg.sender;
   }

    // postObj is a contract function that fires the event logging
    function postObj(string memory hash, string memory rpa_chn_ref, string memory cust_ref) public {
        require(owner == msg.sender, "calling address must be the contract owner/EOA");
        require(bytes(hash).length != 0, "error: empty hash value");
        require(bytes(rpa_chn_ref).length != 0, "error: empty rpachain reference value");

        emit LogVal(hash, rpa_chn_ref, cust_ref, hash, rpa_chn_ref);
        return;
    }

   // LogVal defines the event to be logged
   event LogVal(
       string indexed log_hash,
       string indexed log_rpa_ref,
       string log_cust_ref,
       string log_hash_str,
       string log_rpa_ref_str
   );
}

