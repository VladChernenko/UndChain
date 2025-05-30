# What is UndChain?

UndChain is a **Layer 1 blockchain** built to power a decentralized cloud platform that is **permissionless, trustless, and service-oriented**.

It is not a fork, not a derivative — it’s a ground-up architecture designed to support:
- Decentralized storage
- Computation
- Access to devices and services
- Economic systems with built-in user protection

UndChain brings together multiple disciplines: networking, cryptography, distributed systems, cloud infrastructure, and economics.

There is **no direct equivalent** to UndChain — it is not simply a Filecoin, Flux, or Ethereum variant. It combines elements of those systems, but in a fully integrated, modular, and unified way.

This makes it both extremely powerful and inherently complex. That’s why we split the project into clear domains.

At the **center** of that system is **UndChain Core** — the protocol layer responsible for all communication, coordination, validation, and shared infrastructure that powers the rest of the network.

---
# Understanding UndChain Core

**Purpose**: This document is being made as a reference guide into how and why certain aspects of the code were written. This is meant to provide insight into why the structure is made the way it is architecturally and how it can interlink together (i.e. big picture)

Each section will be separated based on a theme or task that is meant to be completed. It should also be accompanied with all of the relevant files for that section. 

This document is a part of a much larger system. The goal is to have each piece of the system be its own self contained system and lays the foundation for how our co-chains will operate on the network.

---
# What is UndChain Core?

UndChain itself is a **layer 1** project that is focused on providing decentralized cloud computing that is permissionless. It is a very advanced and novel system that comprises of many systems working together in order to make this happen. One of the greatest difficulties in this system is making it trustless. Nothing compares to UndChain as nothing exists to compare it to, we are in a league of our own which places us at a significant lead over anyone wanting to make a system like this. This project has been split between several disciplines with the UndChain core team being at the center of that system; its the foundation that will drive all systems (which we call co-chains).

UndChain core comprises of all of the fundamental structures on UndChain and defines how the network operates. It defines 

- Structure for all user types (Validators, Clients, Chain Owners and Partners). 
	- This includes how each are meant to contact one another (networking protocols)
- It also defines base protocol elements, this includes things like our asset protection systems (Will, Freeze and Limiter)
- defines our standard asset contact types. 
- How encryption works on chain
- Defines the perception score (reputation system) and establishes the algorithm for loss and gain.
- Defines base messaging system that allows users to communicate with one another on chain
- Defines our scalability protocols and when to create subdomains (mixture of latency and quantity of users)
- Defines consensus types and the receipt system
	- Storage - Takes 4GB chucks of storage and uses those as 'containers' across the network for user to fill with data
	- Computation - Splits tasks among multiple partners who then are issued standard challenges during execution to determine if they are computing data or not
	- Access - Uses witnesses to determine if an event occurred and shows proof of those events
- Maintains metrics on the network
	- Analytics
	- List of Co-chains
- Defines payment systems
	- Subscription models
	- Redemption system
	- Return policy
- Defines core protocols
	- Convergence protocol - meant to keep the size of the blockchain small
	- UnaS - System that maps user names to public keys

---

# Validator.py Operation

This section goes over how a validator is supposed to interact with one another (does not include interactions between other user types). Validators go through stages on initialization in order to sync with the group. Found in `validator.py` we can see that a validator has multiple states:

```Python
class ValidatorState(Enum):

    DISCOVERY = 1
    SYNC = 2
    PENDING = 3
    REDIRECT = 4
    ACTIVE = 5
    ERROR = 6
```

Each state defines a moment in time as the validator is initializing, we start in discovery. This is where we are actively searching for a network or if we are a known validator for the network we want to work for we only reach out to the other known validators. 

- **Discovery** - This happens when the validator first initializes and doesn't end until it finds another validator that is NOT in discovery mode.
- **Sync** - This mode is reserved in the event that a validator has found another validator that is in the active state and has made a request to sync it files with the active validator. 
- **Pending** - This state is reserved when a validator has been picked to serve in the active validator pool and is preparing to go live
- **Redirect** - This happens when the validator is not active (meaning they are not participating in the pool due to the hard limit in number of validators). This was created so that if another user accidently contacts them they can redirect that use to the active validators.
- **Active** - This is reserved for validators who are actively accepting job requests on the network and directing traffic by matching partners and clients.
- **Error** - This state is reserved for any errors that are encountered during this process. Should be set inside every error handler found in the validator. 

```Python
def __init__(self, public_key: bytearray, rules_file: str) -> None:

        logger.info("Initializing Validator")

        self.state = ValidatorState.DISCOVERY

        self.public_key: bytearray = public_key

        self.run_rules = RunRules(rules_file)

        logger.info(f"Rules for {rules_file} have been loaded")

        self.run = False

        self.is_known_validator: bool = self.check_if_known_validator()

        self.comm: AbstractCommunication

  

        self.packet_generator = PacketGenerator("2024.09.30.1") # Get the version from the run rules file

        self.packet_handler = PacketHandler(self.packet_generator)
```

In the above code we see that 
- we are setting our current state into Discovery (states can be pinged from other validators so they know at what stage the validator is at), 
- we then pull our public key in as this can be shared with others upon request, 
- then we load the run rules file which can be thought of as a configuration file. At this stage the reason its needed is you need the routing information for the known validators. 
- We then set run to false as this is used later inside of the listener and stop methods. Code is async so we need to know when to stop or if it's running.
- Then we check to see if we are a known validator this was used as a conditional to diverge the path between known and unknown validator, but what I found was they both use the same methods (they both reach out to known validators and attempt to get added to the active validator list). I decided to leave this in as there may be a reason I need it in the future, if not we need to delete it. 
- We then call the abstract communication class which was designed this way to allow various mediums of communication. The most common method will be TCP/IP, but the goal is to expand the network to operate outside the traditional internet so a more general approach was needed here. 
- Then we move to the packet generator which is responsible for forming the packets that are meant to be sent and received on the network (currently only focused on validators). These packets form a small header at the beginning so they can be quickly identified as to what type of message this is. The variable passed in is the version number of the communication protocol being used. All versions on UndChain follow a `YYYY.MM.DD.x` format where Y = Year, M = month, D = day and x is any subversion that may exist 9in the event that a co-chain need to have a hot patch for a security flaw). This should be pulled from the run rules file and not statically assigned as I have done. 
	- Reasoning behind version is that if we have to expand our communication protocol a receiver can look at the version and tell if they can interpret what is being sent. If not then they send a message back to the sender with a version that they are using and depending on who is older request an update to the new an updated version. 
	- *This is not implemented yet as it doesn't even pull the version from the run rules file.* 
- Lastly we have the packet handler which is used in our async functions later to handle any packets coming through the listener. Specifically it decodes what those packets are and takes actions based upon the packet type. 

```Python
    async def start_listener(self) -> None:

        '''

        This method is responsible for setting up and running

        the listener portion of the validator until it's terminated.

        '''

  

        logger.info("Starting validator listener...")

  

        # Start the listener in the background

        try:

            self.comm: AbstractCommunication = CommunicationFactory.create_communication("TCP")

        except ValueError as e:

            logger.error(f'Fatal error. Unknown communication type: {e}')

            self.state = ValidatorState.ERROR

            raise ValueError(e)

        # Need to grab our real IP info later

        asyncio.create_task(self.comm.start_listener("127.0.0.1", 4446))

  

        while self.run:

            message: bytes = await self.comm.receive_message() # Get the message

            await self.handle_message(message)
```

The start listener section is built so that validators can active listen for TCP connections coming in (for now its only other validators, but when the system is operational is will be any user type). This was made async so that we would only execute in the event that we receive a packet in and ideally would scale across multiple cores as at this time this is not optimized to handle a large number of requests (last test I ran clocked in at 10k). If you notice we are manually calling start listener on localhost, but in the actual system we would want to run on *this computers IP*, so that its discoverable to external machines. *I believe 0.0.0.0, should do this...*

Once a message is received it goes to `handle_message`, which simply takes the message, and sends it to the packet handler for processing. Handle message was placed here because it will later condition the message prior to handing it off to the packet handler, by decrypting the message being sent using the validators private key. *Most communication on UndChain is encrypted, matter a fact the only time it isn't is when a user is requesting the public key from another user.* At the time of writing this functionality has not been built in to aid in debugging. 

```Python
async def stop(self) -> None:

        '''

        This method is responsible for ending the validator loop

        and to communicate to it's peers that it is going offline

        '''

  

        self.run = False

        logger.info(f"shutting down the validator.")

  

        try:

            await self.comm.disconnect() # type: ignore

            logger.info(f'Successfully stopped listening')

        except Exception as e:

            logger.error(f'Failed to stop validator from listening. Inside Validator:stop()')
```

This method is used to inform other validators that it's about to go offline, this is done because Validators that simply go off line with no notification receive a lowed perception score. While they will still encounter a lower perception score, by going offline regardless if they present this or not, it will be less sever. It also gracefully shuts down the listener, which is important for memory management and security. *Note: at this time the code does NOT send a notification to other validators, that needs to be added in*

```Python
def set_state(self, new_state: ValidatorState) -> None:

        '''

        Changes the state of the validator which is used to determine

        this validators readiness on the network.

        '''

  

        logger.info(f"Transitioning to {new_state.name} state.")

        self.state: ValidatorState = new_state
```

Simply sets the state of the validator as it's progressing through it roles (identified in the ENUM above)

```Python
async def handle_message(self, message: bytes) -> None:

        '''

        Send message over to the packet handler for processing.

        '''

  

        logger.info(f"Handling message: {message}")

  

        try:

            response: None | bytes = self.packet_handler.handle_packet(message)

  

            if response:

                await self.comm.send_message(response, bytearray(b'recipient_public_key')) # Need to get teh public key of who we are sending this to

                logger.info("Response sent back to sender")

            else:

                logger.warning("No response sent back for this packet type")

  

        except Exception as e:

            logger.error(f'Failed to process message: {e}')
```

This was referenced earlier, but is responsible for taking in a message coming from the listener and routing it to the appropriate packet handler. If there is no handler, we should return and error stating that we could not process the message.  

```Python
def send_state_update(self, recipient: bytearray) -> None:

        '''

        This method is used for appending the validators state to the

        beginning of a incoming request so that the user knows the heath

        status of this validator

        '''

  

        state_info: LiteralString = f"State: {self.state.name}"

        logger.info(f"Sending state update to {recipient.decode('utf-8')}: {state_info}")

        # Logic to send the state update

        ...
```

When I made this method originally it was when the validator had controlled all packet handling (prior to creating packet handler). Since then I have kept it in as I was thinking that this could be used in the header of each message as a way of consistently providing the state of the validator without a user explicitly asking for it. The method signature would need change as we are no longer sending in a message and we would be returning the state... Probably need to just make this a getter for the state...

```Python
def handle_error(self, error_message: str) -> None:

        '''

        Logic to handle errors and transition to the validator

        into the ERROR state. Validator should communicate this state

        to it's peer (other validators in the pool).

        '''

  

        logger.error(f"Error occurred: {error_message}")

        self.set_state(ValidatorState.ERROR)

        # Implement recovery or notification logic here

        ...
```

This is a place holder for now, but the intent is that if we receive an error we execute this method with the idea it would gracefully handle the error. At minimum, logging the error. ideally by returning a correction for that specific error. We could handle errors on a case by case basis too, so this may not be needed. 

```Python
async def discover_validators(self) -> None:

        '''

        This method is for discovering other validators or listening for

        incoming requests for validators to join the pool.

        '''

  

        logger.info("Discovering validators asynchronously...")

  

        known_validators: list[str] = self.run_rules.get_known_validator_keys()

        tasks = [] # Collect tasks for connecting validators

  

        for validator_key in known_validators:

            if validator_key == self.public_key.decode('utf-8'):

                logger.info(f'You are a known validator {validator_key}, so we are not contacting ourselves')

                continue

  

            logger.info(f'Attempting to connect to validator: {validator_key}')

  

            # Get contact info for this validator

            contact_info: dict[str, str] = self.get_contact_info(validator_key)

            if contact_info:

                try:

                    logger.info(f'Initializing communication with {validator_key} using {contact_info["method"]}')

                    try:

                        comm: AbstractCommunication = CommunicationFactory.create_communication(contact_info["method"])

                    except ValueError as e:

                        logger.error(f'Fatal error. Unknown communication type: {contact_info["method"]}')

                        self.state = ValidatorState.ERROR

                        raise ValueError(e)

                    tasks.append(self.connect_to_validator(comm, validator_key, contact_info))

                except Exception as e:

                    logger.error(f'Failed to connect to validator {validator_key}: {e}')

            else:

                logger.error(f'Failed to retrieve contact info for validator {validator_key}')

  

        # Await all of the gathered tasks

        if tasks:

            await asyncio.gather(*tasks)

        else:

            logger.info("No other validators to connect to...")
```

This is the algorithm that we use to discover other validators on the network. Inside here we request the known validators inside of a list; we then create an empty task list, that is designed to later parallelize sending requests out (speed up the process in the event that you have many known validators). 

We then check to see if we are a known validator so we don't try to contact ourselves. If we are not the validator we are currently looking at we continue on by attempting to contact them using the info contained in the `[route]` section of the run rules file. This is done using the method `get_contact_info`, if we receive that info then we try communicating using that route. *Remember that currently we have only implemented TCP/IP, but UndChain is designed to work across multiple communication types (think Bluetooth or LoRA) so we call the `communicationFactory` to figure out how to do it.* I am not sure if we should set the error for that as a fatal error or just simply proceed to the next validator in the list. 

The last potion of this code takes all the rout information that we have gathered for each validator and places them in a pool to be executed all at asynchronously. 

```Python
async def connect_to_validator(self, comm: AbstractCommunication, validator_key, contact_info):

        try:

            await comm.connect(bytearray(validator_key, 'utf-8'), contact_info) # type: ignore

        except Exception as e:

            logger.error(f'Failed to connect to validator {validator_key}: {e}')
```

This is the helper function used in `discover_validators` mentioned above. The attempts a TCP connection to the end device. 

```Python
def get_contact_info(self, public_key: str) -> dict:

        '''

        Retrieves the contact information for a validator from the run rules

        based on the public key being passed in.

  

        Returns:

            Dictionary with the type of communication and the route

        '''

        known_validators = self.run_rules.get_known_validators()

  

        for validator in known_validators:

            if validator['public_key'] == public_key:

                logger.info(f'Found contact info for public key {public_key}: {validator["contact"]}')

                return validator['contact']

        raise ValueError(f'Validator with public key {public_key} was not found in the rin rules file.')
```

This is another helper method for `discover_validators`, its job is to take a public key and find the contact information (aka route information) for the particular validator so the system knows how to contact them.  

```Python
def check_if_known_validator(self) -> bool:

        '''

        This method is responsible for determining if this validator is

        apart of the known validator class within this co-chain

        '''

        known_validator_keys: list[str] = self.run_rules.get_known_validator_keys()

        public_key_str: str = self.public_key.decode("utf-8")

        is_known: bool = public_key_str in known_validator_keys

        return is_known
```

This helper function simply checks to see if the current profile is a known validator, based on the run rules list (you can also think of this as a configuration file for a co-chain).

```Python
if __name__ == "__main__":

    async def main() -> None:

        public_key = bytearray("validator_pub_key_3", "utf-8")

        run_rules_file: str = "UndChain.toml"

        validator = Validator(public_key, run_rules_file)

  

        try:

            await validator.start_listener()

            await validator.discover_validators()

  

            while validator.run:

                await asyncio.sleep(1)

  

        except ValueError as e:

            logger.error(f'May need to check the run rules file: {run_rules_file} \nThere is a misconfigured communication type')

            return # End program to prevent undefined behavior. TODO: Create a checker to see where in the TOML file we have the misconfiguration.

        finally:

            print("System listening for new connections...")

            await validator.stop()

  

    asyncio.run(main())
```

At the end of nearly all Python modules, I add a small test at the end to ensure it works as intended. This scripts job at this time is to simulate finding validators based on the routes contained within the run rules file. If no connections can be made, it terminates the connection (via TCP timeout). If there is no end device you will see an error code like this:

```CMD
[ERROR]  - Failed to connect to validator validator_pub_key_1: [WinError 1225] The remote computer refused the network connection
```

---

# Run Rules

The `run_rules.py` file is meant to create a well structured system for taking in the run rules file for a specified co-chain and interpreting those rules so that validators (and later partners) can interpret how the chain is to operate / function. This script will need to be extended and there are some methods that are simply not implemented yet, but will need to be. The run rules file can be thought of as the closest UndChain will get to a 'smart contract' as this establishes ideas such as tokenomics and sets preferred validators. This is the entry point for any blockchain project that enters into UndChain. 

```Python
def __init__(self, config_filename: str) -> None:

        # Construct the path to the run rules file

        root_dir: str = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))  # Navigate to the root directory

        run_rules_path: str = os.path.join(root_dir, 'Run Rules', config_filename)

  

        # Load the TOML file

        with open(run_rules_path, 'rb') as f:

            self.config: Dict[str, Any] = tomllib.load(f)
```

This code defines where to look for the run rules files will exist (currently located a directory above the script itself) and based on which file is selected (co-chain), it loads the rules into a dictionary which is used to parse / extract through the remainder of the code. 

Note: When the UI system (M3) is implemented users should be able to select which co-chains they wish to support. They will at that time be able to make a tier list of which they would like to support so that even when one validator pool is full they are able to support other co-chain validator pools. Users will be able to download addition run rules files as they subscribe to the co-chain. 

**Future addition**: We need to add an error handler here so that in the event a run rules file is requested that doesn't exist the program doesn't crash, while also notifying what called this method what happened so it can handle this error accordingly.

```Python
def get_job_file_structure(self, co_chain_name: str = "base_job_file") -> Dict[str, Any]:

        '''

        Fetch the job file structure for the base job file or a specific co-chain.

        '''

  

        job_structure: Dict[str, Any] = {

            "fields": self.config[co_chain_name]["fields"],

            "mandatory": self.config[co_chain_name]["mandatory"],

            "job_types": self.config[co_chain_name]["job_types"],

            "token": self.config[co_chain_name]["token"]

        }

        return job_structure
```

This section of code pulls out the job structure fields within the run rules file. It attempts to pull out the values inside fields:

```TOML
[base_job_file]

fields = ["user_id", "job_type", "min_partners", "block_id", "block_time", "job_priority"]

mandatory = ["user_id", "state", "block_id"]

job_types = ["transfer", "auction", "naming_service", "store_req", "dmail"]

token = "UGP"  # The token used for transactions
```

- **Fields** - This defines all of the various fields that can be sent to the validator during a request. For example, a validator could request the current block time which is important during a time sync. 
- **Mandatory** - This defines fields that MUST be provided by a user when responding to any request. In this case, we must always so who we are, what status we are in (Validator state) and what block ID we are currently processing. 
- **Job Types** - This field can be thought of what methods can be called from this chain. For example a user could request the store command in order to store data on the network.
- **Token** - Directs which token is needed to perform this function, this is specific to partners as validators can only accept UGP. If not provided we should always assume USP. 

The methods and fields listed are NOT final and will change as the system evolves, for example there is currently no command for our asset protection protocols at this time. 

I believe that we should have error handling for missing (or empty) fields as that could happen if the file is corrupted someway on the end users computer. Think we should also have a hash that can be used to check against the network to ensure this file wasn't tampered with for security purposes.

```Python
def get_validator_info(self) -> Dict[str, Any]:

        """

        Fetch the validator information including max and known validators.

        """

  

        max_validators = self.config["max_validators"]["max"]

        known_validators = self.config["known_validators"]

        return {

            "max_validators": max_validators,

            "known_validators": known_validators

        }
```

Get validator info is meant for collecting the list of known validators from the run rules file, as well as what the maximum number of validators can exist on a network. This will give developers some flexibility to define what sort of network they want. The could go with less Validators which means they will reach consensus faster, but it may not have enough bandwidth for high throughput systems. 

Even if a run rules file has more validators that the max 44% that the network allows, the network shall ignore those 'extra' validators and place them in an inactive state with the idea that they can be added if one of the others goes down. We must adhere to having no more that 44% of known validators on a co-chain as that increases concerns with centralization. 

```Python
def get_utilities(self) -> Dict[str, Any]:

        """

        Fetch the list of utilities available on the chain.

        """

        return self.config.get("utilities", {})
```

This is one of those methods that I have yet to implement. The idea is that we can get a list of utilities that the partners can perform. Each utility should have a suggested fee (partners will be able to set their own fees), what the name of the method is and a description of what that method does.

```Python
def get_sub_domain_info(self) -> Dict[str, Any]:

        """

        Fetch the sub-domain information including linked co-chains.

        """

        return self.config.get("sub_domains", {})
```

This method (unimplemented), shall pull the sub-domain information for the co-chain (in this case UndChain Core). Sub-domains are apart of UndChain's scaling solution. In short when the network notes a significant amount of lag between users and validators it will create a subdomain. This ensures that the response times between all users is fast and responsive. All sub-domains will report back to the primary domain so that things such as transaction history are synced at all times. 

Users will be able to select their sub-domain, but under normal operation a user will attempt to connect to a sub-domain and if the latency is too high another sub-domain will provided that they can try. 

Just like the main chain; chain owners can set known validators for these sub-domains as a means of ensuring the network is operating as intended. Chain owners can also elect to NOT allow a sub-domain to be created as I can see this as a possible attack vector that will lure unsuspecting clients in. If over time a subdomain simply isn't being used it will be merged back with an adjacent sub-domain. 

```Python
def get_governance_rules(self) -> Dict[str, Any]:

        """

        Fetch the governance rules such as voting period and quorum.

        """

        return self.config.get("governance", {})
```

This fetches the governance rules for a particular co-chain. In short if a co-chain uses a voting system to decided upon tokenomics or features of the network then this can be used as a standard structure for deciding qualifying factors. This includes:

- Age of the account
- Users perception score
- fee associated with voting 
- Timing window - Defines when and how long voting is active
- Network Usage - Meaning how often does a user interact with that co-chain
- access list that are set by the chain owner - *honestly not a fan of this one, but we are making a platform that anyone can use and at least you can see who is in control*

As an example, USP will have a voting mechanism where users can select to either halve or double the supply of tokens every cycle (each cycle is four years in length).

```Python
def get_tokenomics_rules(self) -> Dict[str, Any]:

        """

        Fetch the tokenomics rules such as token issuance and payout timing.

        """

        return self.config.get("tokenomics", {})
```

This extracts the tokenomics of the system which include:

- **Startup** - This defines how many tokens were generated at the beginning of the co-chain. *For example USP will mint 44,444,444 at it's genesis*
- **Emissions** - How many tokens are supposed to be emitted
- **Share** - Defines how tokens are to be distributed. *My thoughts are it's either equally distributed **or** its based off of work load*
- **Timing** - Defines how often this payout occurs
- **Fee** - This defines transaction fees that may occur with this token; both sending and receiving. This is based in percent. You will have a base fee for transactions and then a breakdown of where that fee goes. *We may add validators as a group, but under normal circumstances they earn network tokens not utility tokens*
- **Burn** - If the co-chain has a burn mechanism this would define how the burn would function. This is a lot like the fee except the recipient is no one. It will be recorded on the network that way too.
- **Stake** - Defines lockup times for any sort of stake options that a token may have along with its reward.
- **Cap** - This goes well with staking as this limits the amount of rewards any one account may receive. Debating if we should make this a progressive system or not. *Must keep code out of the config file.* 
- **Vesting** - This defines how long a token must be held for stake rewards. Should help a project in major sell offs. 
- **Bridge** - This one wont be useful until we have a bridge system working, but the idea is that this section will contain the lockup amount of tokens on the network. NOTE: This a location, not a value as this will fluctuate a lot we would be updating this file far too often to just set the value. 

This is a massive feature for UndChain as it allows anyone to easily look at the structure of a token and know exactly how this token behaves, without needing to audit a smart contract. I believe these are the only tokenomic systems that are truly out there, but they could expand if there are any new systems in the future. 

Example of what this *could* look like:

```TOML
[tokenomics]
# Genesis mint
startup = 44444444

# Emission system
emissions = 4444          # Total emissions
timing = "daily"          # Distribution cycle (e.g., monthly)
share = "workload"          # Can be "equal" or "workload"
frequecy = 4              # States how often payout occurs in hours

# Global fee applied to relevant tx types
[fee]
percent = 0.02              # 2% total fee on applicable transactions

# Fee distribution (must sum to <= percent above)
[[fee.distribution]]
percent = 0.004
to = "partner"              # System keyword for partner rewards

[[fee.distribution]]
percent = 0.010
to = "validator"            # System keyword for validator reward bonus

[[fee.distribution]]
percent = 0.003
to = "@UndChain Treasury"   # UnaS-registered public name

[[fee.distribution]]
percent = 0.003
to = "@DevFund"             # Another registered name (could be multisig DAO)

# Burn settings (optional deflation)
[burn]
on_transfer = 0.005         # 0.5% of each transfer is burned
on_redemption = 0.01        # 1% burned on redemption events

# Staking logic
[stake]
lockup_days = 30            # Required stake lock-in period
reward_rate = 0.07          # APY (or proportional over timing cycle)

# Optional cap to limit max rewards per account
[cap]
enabled = true
max_rewards_per_account = 50000

# Optional vesting schedule
[vesting]
enabled = true
cliff_days = 90
linear_release_days = 365

# Bridge lockup placeholder (future use)
[bridge]
location = "@undchain/ethBridge"

```

Should have mentioned earlier, but we will want to check these files prior to deployment. In this example we would want to check and ensure that the split in fees equal 100% and flag for anything over or under. *NOTE: Timing for blocks will be in upcoming section*

```Python
def get_performance_metrics(self) -> Dict[str, Any]:

        """

        Fetch the performance metrics like max block time and latency thresholds.

        """

        return self.config.get("performance", {})
```

This section pulls the configuration needed to set the network’s block timing and latency thresholds, which are used to determine when a subdomain should be created. While this information is primarily consumed by validators, it’s publicly visible so any user can audit the performance parameters of the chain.

In addition, this config defines how often validator rotation should occur — a key feature for ensuring decentralization by giving new users a chance to become validators, provided they meet network minimums and stay in the queue.

We also define the maximum number of users allowed per subnetwork and introduce a cooldown timer to prevent subdomains from spawning too aggressively. This avoids edge cases where a few users with poor latency trigger unnecessary subdomain creation simply by connecting more efficiently to one validator.

_Thought: It may be worth designing an M3 widget that displays these parameters in real time, offering transparency into performance targets and validator load._

```Python
def get_subscription_services(self) -> Dict[str, Any]:

        """

        Fetch the subscription services if they are defined.

        """

  

        return self.config.get("subscription_services", {})
```

If a chain has a subscription based service on it's **utility** (which the main chain does not), then that can be entered here along with durations of the offer as well as any limits that are associated with it. It also allows a chain owner to decide **if** they will permit refunds and what those look like. 

We shall also have the following fields so that chain owners can further customize their offerings. 

- **Price** - Cost per each subscription tier and what services that provides. The simplest is the All subscription which is what the network token uses. It means that any function(s) called by the user are covered in the subscription service. 
- **Duration** - How long does this subscription last
- **Renewable** - Can users auto renew or not, double edge sword as if you auto-renew then you have guaranteed revenue, but if you later change the cost those users who set this up will still receive it at the same price...
- **Limits** - Sort of defeats the purpose of a subscription, but you can set limits to each of the services you provide
- **Throttle** - Like limit except it reduces the responsive time of the request
- **Grace** - How long you give the customer before they are kicked from the program for non-payment. Losing access is immediate, this is if they lose the price they are currently paying for the service. 
- **Refund** - Define the elapsed time for a refund (in hours) as well as the amount that will be funded in percent. 
- **Description** - This helps with UI and gives the creators an opportunity to provide a brief on how pricing is managed.

*IMPORTANT: The network token (UGP) does have a subscription system, but that will be hard coded. Remember that these rules apply to co-chains, NOT the main network*

```Python
def validate_job_file(self, job_data: Dict[str, Any], co_chain_name: str = "base_job_file") -> bool:

        """

        Validate a job file against the mandatory fields for a specific co-chain.

        """

  

        mandatory_fields = self.config[co_chain_name]["mandatory"]

        return all(field in job_data and job_data[field] is not None for field in mandatory_fields)
```

Not sure if we will keep this here or not, the idea is that we want to make sure that this is a properly formed run rules file. What I am thinking at this time is that it will be a utility inside code ledger. Sort of like a test in how we validate and compile code when we implement Pseudo. 

```Python
def get_known_validator_keys(self) -> list[str]:

        '''

        Retrieves a list of all the known validators from the run

        rules file.

        '''

  

        known_validators = self.config["known_validators"]

        return [validator["public_key"] for validator in known_validators]
```

This is to provide users the ability to encrypt messages that are being sent to the known validators that are listed. This is done to ensure that communications between entities are kept private the entire time. 

```Python
def get_known_validators(self) -> list:

        return self.config["known_validators"]
```

This method pulls all of the known validators inside the run rules file. This is used to parse through when validators are being discover. It also assist other users the ability to contact the validator pool. 

```Python
def get_min_validator_score(self) -> int:

        '''

        Obtain the minimum validator perception score required to join

        the network

        '''

  

        score = self.config.get("min_validator_score", 0)

        if isinstance(score, int):

            return score

        else:

            logger.warning(f"'min_validator_score' is not an integer. Returning default of 420.")

            return 420
```

Returns the minimum perception score for validators to be able to run on the network. Defaults to 420.

Shouldn't we make a method for performance?

```Python
def get_min_partner_score(self) -> int:

        '''

        Obtain the minimum validator perception score required to join

        the network

        '''

  

        score = self.config.get("min_partner_score", 0)

        if isinstance(score, int):

            return score

        else:

            logger.warning(f"'min_validator_score' is not an integer. Returning default of 420.")

            return 420
```

Exactly the same as the validator score requirement, just changed for the validator. I decided to keep these separate as validators should be held to a higher standard, but this gives chain owners more of an opportunity to customize what their platform needs.

```Python
if __name__ == "__main__":

    run_rules = RunRules("UndChain.toml")

  

    # Print a list of known validators from the run rules file

    known_validators = run_rules.get_known_validators()

    print(f'Known validators: {known_validators}')

    # Fetching job file structure for the base job file

    job_file_structure: Dict[str, Any] = run_rules.get_job_file_structure()

    print("Job File Structure:", job_file_structure)


    # Fetching validator information

    validator_info: Dict[str, Any] = run_rules.get_validator_info()

    print("Validator Info:", validator_info)


    # Fetching utilities available on the chain

    utilities: Dict[str, Any] = run_rules.get_utilities()

    print("Utilities:", utilities)


    # Fetching sub-domain information

    sub_domain_info: Dict[str, Any] = run_rules.get_sub_domain_info()

    print("Sub-Domain Info:", sub_domain_info)
  

    # Fetching governance rules

    governance_rules: Dict[str, Any] = run_rules.get_governance_rules()

    print("Governance Rules:", governance_rules)


    # Fetching tokenomics rules

    tokenomics_rules: Dict[str, Any] = run_rules.get_tokenomics_rules()

    print("Tokenomics Rules:", tokenomics_rules)


    # Fetching performance metrics

    performance_metrics: Dict[str, Any] = run_rules.get_performance_metrics()

    print("Performance Metrics:", performance_metrics)


    # Fetching subscription services

    subscription_services: Dict[str, Any] = run_rules.get_subscription_services()

    print("Subscription Services:", subscription_services)


    # Example job data for validation

    job_data: Dict[str, str] = {

        "user_id": "user123",

        "job_type": "transfer",

        "block_id": "0001",

        "block_time": "2024-08-30T12:34:56Z",

        "job_priority": "high"

    }


    # Test retrieving minimum scores

    min_validator_score = run_rules.get_min_validator_score()

    min_partner_score = run_rules.get_min_partner_score()

    print(f"Minimum Validator Score: {min_validator_score}")

    print(f"Minimum Partner Score: {min_partner_score}")

    # Validate the job file

    is_valid: bool = run_rules.validate_job_file(job_data)

    print(f"Is job file valid? {is_valid}")
```

Tests each method against `UndChain.toml` to test and make sure this class functions as intended.

## Conclusion

The run rules file is essential to creating easy to read, contract information that goes into detail into how each co-chain functions. This section has introduced a lot of core concepts that the network will need in order to operate. What's important to not again is that the run rules file only deals with utility tokenomics and NOT the network tokens. All co-chain validators receive UGP as that is our network token. *The network tokenomics section has yet to be defined.*

---

# Packet Handler

The idea of `packet_handler.py` is that its a centralized portion of the code that is dedicated to taking in a packet and then determining what method(s) need to be ran in order to handle that packet appropriately. You can think of the packet handler much like a mail sorter, it looks at the address (the packet header) and from there determines where the content is supposed to go. 

```Python
def __init__(self, packet_generator: PacketGenerator) -> None:

        '''
        Initialize the packet handler
        '''

        self.packet_generator: PacketGenerator = packet_generator

        self.handlers = {

            PacketType.VALIDATOR_REQUEST: self.handle_validator_request,

            PacketType.VALIDATOR_CONFIRMATION: self.handle_validator_confirmation,

            PacketType.VALIDATOR_STATE: self.handle_validator_state,

            PacketType.VALIDATOR_LIST_REQUEST: self.handle_validator_list_request,

            PacketType.VALIDATOR_LIST_RESPONSE: self.handle_validator_list_response,

            PacketType.LATENCY: self.handle_latency,

            PacketType.JOB_FILE: self.handle_job_file,

            PacketType.PAYOUT_FILE: self.handle_payout_file,

            PacketType.SHUT_UP: self.handle_shut_up,

            PacketType.CONVERGENCE: self.handle_convergence,

            PacketType.SYNC_CO_CHAIN: self.handle_sync_co_chain,

            PacketType.SHARE_RULES: self.handle_share_rules,

            PacketType.JOB_REQUEST: self.handle_job_request,

            PacketType.VALIDATOR_CHANGE_STATE: self.handle_validator_change_state,

            PacketType.VALIDATOR_VOTE: self.handle_validator_vote,

            PacketType.RETURN_ADDRESS: self.handle_return_address,

            PacketType.REPORT: self.handle_report_packet,

            PacketType.PERCEPTION_UPDATE: self.handle_perception_update_packet,

        }
```

Inside the init method we set our packet generator (which will be needed when we need to make a response to a user query) we then identify a master dictionary that takes in the type of packet and sets callers to where each packet type will go to be processed. Many of these packet types have yet to be fully defined so a lot of work needs to be done in this file so that we are not only interpreting the packets, but also preforming the correct actions. I chose this method as this is will run at O(1) dispatch via hash map ensures fast routing across dozens (or hundreds) of packet types. 

*NOTE: This is NOT a complete list of all of the available packets as we have not implemented anything for any other user type outside of validators.* 

```Python
def handle_packet(self, packet: bytes) -> Optional[bytes]:

        '''

        Receives a packet, decodes it, and calls the appropriate handler.

        Returns a response packet if needed, otherwise None.

        '''

        try:

            # Extract the first 4 bytes for version and 8 bytes for timestamp (keep this data for later use)

            version_data = struct.unpack('!HBBB', packet[:5])  # 16-bit year, 8-bit month, day, subversion

            timestamp = struct.unpack('!Q', packet[5:13])[0]  # 64-bit timestamp

            # Store version and timestamp for later use (or logging)

            self.version_info = {

                "year": version_data[0],

                "month": version_data[1],

                "day": version_data[2],

                "sub_version": version_data[3],

                "timestamp": timestamp

            }

            # Convert the timestamp to a human-readable format

            human_readable_timestamp: str = datetime.utcfromtimestamp(timestamp).strftime('%Y-%m-%d %H:%M:%S UTC')

            logger.info(f"Packet version: {self.version_info}")

            logger.info(f"Packet timestamp: {human_readable_timestamp}")

            # Now that the version and timestamp are stripped, identify the packet type

            packet_type_value = struct.unpack('!H', packet[13:15])[0]

            packet_type = PacketType(packet_type_value)

  

            logger.info(f"Received packet of type: {packet_type.name}")

            # Debugging: Print payload in hex format for comparison

            logger.debug(f"Packet payload: {packet[15:].hex()}")

            # Now pass the remainder of the packet (payload) to the handler

            handler = self.handlers.get(packet_type)

            if handler:

                return handler(packet[15:])  # Pass the payload

            else:

                logger.error(f"Unknown packet type: {packet_type}")

                return None

        except Exception as e:

            logger.error(f"Failed to handle packet: {e}")

            return None
```

This method takes in a raw packet and then strips out the version information as well as the packet type. It then decodes the type of packet it is (also contained in the header) and calls `handlers.get()` so that it can direct what to do with the data that's inside the packet.  

Just for clairity in the above code:

```Python
version_data = struct.unpack('!HBBB', packet[:5])  # 16-bit year, 8-bit month, day, subversion

timestamp = struct.unpack('!Q', packet[5:13])[0]  # 64-bit timestamp
```

|Code|Stands For|Size (bytes)|Description|
|---|---|---|---|
|`!`|Network byte order|—|Big-endian (used in networking)|
|`H`|**Unsigned short**|2|16-bit unsigned integer|
|`B`|**Unsigned char**|1|8-bit unsigned integer|
|`Q`|**Unsigned long long**|8|64-bit unsigned integer|


```Python
def handle_validator_request(self, packet: bytes) -> Optional[bytes]:

        '''

        Handles an incoming validator request packet and returns a confirmation packet.

    A validator request packet is sent when a validator wishes to join the active pool.
    This method extracts the validator’s public key and responds with a confirmation packet.

    Note: Currently, the public key is logged but not yet added to any internal validator state list.

        '''

        logger.info("Handling Validator Request")

        # Unpack the public key (the packet type has already been stripped)

        try:

            public_key = packet.decode("utf-8")  # The entire payload is the public key

        except Exception as e:

            logger.error(f"Failed to unpack the packet: {e}")

            return None

        logger.info(f"Validator request from: {public_key}")

        # Generate a response using packet generator

        confirmation_packet: bytes = self.packet_generator.generate_validator_confirmation(position_in_queue=4)

        return confirmation_packet
```

This section defines what happens when a validator request packet has been sent. A validator request packet is sent when a validator wishes to become part of the validator pool, at the time of genesis this will be executed to each of the known validators listed in the run rules file so that they can be added and those known validators will sit in a pending state until they hear from six more unknown validators prior to going into the next phase which is **time sync**. 

Currently this methods only goal is to extract the public key from the file so that it can be added to a validator list, inside that list we should note the status of the validator and what their performance score, perception score and position in queue. *NOTE: That list is not implemented yet it simply prints the public key and queue position.* Lastly a confirmation that the packet was received and it was added to the list is sent. 


```Python
def handle_validator_confirmation(self, packet: bytes) -> None:

        '''

        Handles validator confirmation packet
        
        This packet is sent in response to a validator request         and contains the validator’s assigned position in the          queue.

        '''

        logger.info("Handling Validator Confirmation")

        # Unpack the queue position directly (since the payload is already stripped)

        try:

            queue_position = struct.unpack(">I", packet[:4])[0]  # First 4 bytes of the payload = queue position

            logger.info(f"Validator confirmed in queue position: {queue_position}")

        except Exception as e:

            logger.error(f"Failed to unpack the packet: {e}")
```

This is called in the response of a validator packet being sent. It's goal is to return the que position of the validator. *Thought: Should we also send the validator list?*


```Python
def handle_validator_state(self, packet: bytes) -> None:

        '''
        Handles validator state packet.
        '''

        logger.info("Handling Validator State")

        # Unpack and log the validator state

        state = packet[2:].decode("utf-8")

        logger.info(f"Validator state is: {state}")

        ...
```

This is not fully implemented but will return the current validator state. 

```Python
def handle_validator_list_request(self, packet: bytes) -> None:

        '''

        Handles a validator list request packet.

    This packet is used to retrieve the current list of validators in the pool.
    The payload contains two parameters:
    - `include_hash` (1 byte): If 1, return the hash of the validator list; if 0, return the full list
    - `slice_index` (4 bytes): Used to paginate or split responses if the list is large

    This method allows validators to sync their view of the network without always attaching the full list to every confirmation packet.

        '''

        logger.info("Handling Validator List Request")

        # Unpack modifiers

        include_hash, slice_index = struct.unpack(">BI", packet[2:7])

        logger.info(f"Validator List Request: Include Hash: {include_hash}, Slice Index: {slice_index}")

        ...
```

Handles a validator list request coming in from the user and (if sent) performs a slice on that list to save on bandwidth (used if the validator is missing some entries in their list). 

```Python
def handle_validator_list_response(self, packet: bytes, slice_index: Optional[int] = None) -> None:

        '''
        Handles validator list response packet
        
       This packet includes either a full or partial list of validators, depending on
    what was requested (via slice index). The payload is expected to be a
    UTF-8 encoded comma-separated list of public keys, beginning at the requested offset.
        '''

        logger.info("Handling Validator List Response")

        # Extract the list of validators from the packet

        validators_data = packet[2:]

        validators_data = packet.decode("utf-8")
		validators = validators_data.split(",")
		
		if slice_index is not None:
		    logger.info(f"Received validator slice at index {slice_index}: {validators}")
		else:
		    logger.info(f"Received full validator list: {validators}")

        ...
```

This method handles the response that will be sent as a result of a list packet. It shall include the contents of the list from the point in which the requester asked for it (don't forget they can request a partial list to save on bandwidth. 

```Python
def handle_latency(self, packet: bytes) -> None:

        '''
        Handles latency packet
        '''

        logger.info("Handling Latency Packet")

        # Extract latency counter and perform latency-related operations

        latency_counter = struct.unpack(">I", packet[2:6])[0]

        logger.info(f"Latency Counter: {latency_counter}")

        ...
```

This packet is meant to handle the latency between two validators, this is important as syncing between validators and keeping a constant connection between each other is key to how the network functions. It also helps determine if we need to spin off a sub-domain. *Thought: Should we make max latency a value that can be set in the run rules file for a co-chain? Feels like something a chain owner would want to test.*

```Python
def handle_job_file(self, packet: bytes) -> None:

        '''
        Handles job file packet
        '''

        logger.info("Handling Job File")

        # Unpack job file data

        job_data = packet[2:]

        logger.info(f"Job File Data: {job_data.decode('utf-8')}")

        ...
```

This is the handler for the 'job file', this will store jobs as they come in from clients and pop them off as they are completed. Jobs are massively important on UndChain and will get hashed as a block is completed. Job files are shared across all validators as they must all agree on the job file in order to hit consensus. *This is an incomplete method and needs to be fully implemented*

```Python
def handle_payout_file(self, packet: bytes) -> None:

        '''
        Handles payout file packet
        '''

        logger.info("Handling Payout File")

        # Unpack payout file data

        payout_data = packet[2:]

        logger.info(f"Payout File Data: {payout_data.decode('utf-8')}")

        ...
```

The payout file is just like the job file and this must also be shared across all validators as this is also part of consensus. The payout file is critical as it is what makes UndChain a 'pool-less protocol'. This means that it doesn't matter how many partners (or validators) you have working for you. You get paid in what you do. No token lottery here. *NOTE: This handles only partner payouts, later when we implement partners you will see the payout file for validators.*

```Python
def handle_shut_up(self, packet: bytes) -> None:

        '''
        Handles shut-up packet
        '''

        logger.info("Handling Shut-Up Packet")

        # Perform logic to pause communication or reduce traffic load

        ...
```

The shut up packet was made in the event that a validator is getting spammed with messages from any type of user. In this case this is meant for other validators that may be requesting too many resources at once. This notifies that validator that they need to back off of requests to this validator for a specified cool down period. This should also be shared to other validators so that if, the other ignores this request it can be reported and perception scores can be reduced. The backoff time should be declared within the return packet (in seconds).

Instead of dropping connections or blocking the peer, UndChain supports **graceful communication backoff** — and this packet represents that protocol-level nudge.

```Python
def handle_convergence(self, packet: bytes) -> None:

        '''
        Handles convergence packet
        '''

        logger.info("Handling Convergence Packet")

        # Extract convergence details

        convergence_time = struct.unpack(">I", packet[2:6])[0]

        logger.info(f"Convergence Time: {convergence_time}")

        ...
```

This is one of UndChain's most powerful methods. The convergence is a process in which we take the existing blockchain and hash it into one a new genesis block. the original blockchain is stored on chain as a means of proving older transactions (since UndChain has network storage this is trivial, we simply send a store request to the validators). 

Convergence should be triggered on two main events:

1. When the block chain becomes too large that it takes too long for new validators to join in. *I am thinking no more than 4Gb*
2. When the hashing algorithm is changed (i.e. when we switch encryption methods)

```Python
def handle_sync_co_chain(self, packet: bytes) -> None:

        '''
        Handles sync co-chain packet
        '''

        logger.info("Handling Sync Co-Chain Packet")

        # Unpack and process the sync co-chain data

        co_chain_id = packet[2:].decode("utf-8")

        logger.info(f"Sync Co-Chain ID: {co_chain_id}")

        ...
```

This is received when a new validator is coming online. This gives a new validator the ability to (in one request) ask for the active validators, the job file, the run rules file, the payout file and the blockchain. A validator will stay in sync until this is completed. *Thought: We could implement a mechanism so that no one validator is responsible for providing this data. Perhaps what we could do is send it in chucks that come from various other validators. I do see this being a function that a passive validator would perform rather than an active one.*

```Python
def handle_share_rules(self, packet: bytes) -> None:

        '''
        Handles share rules packet
        '''

        logger.info("Handling Share Rules Packet")

        # Process rule sharing

        rule_version = packet[2:].decode("utf-8")

        logger.info(f"Share Rules version: {rule_version}")

        ...
```

This is critical as validators need a way to share the run rules file of the co-chain they are operating on in a decentralized way. It can come in two forms; one a direct share from fellow validators (which will be the fastest) and two from partners as all co-chains will save their run rules state on the network. *Thought: Could use the network as a checker to ensure you received a valid run rules file rather than downloading the whole file from partners*

```Python
def handle_job_request(self, packet: bytes) -> None:

        '''
        Handles job request packet
        '''

        logger.info("Handling Job Request")

        job_data = packet[2:].decode("utf-8")

        logger.info(f"Job Request Data: {job_data}")

        ...
```

This method is specific in handling the job requests that are coming in from clients (or any user). These requests are placed in the job file upon receipt so that partners can view that file and respond based upon their availability and ability (to provide the requested service). *NOTE: I believe we should also have an algorithm that allows partners to 'subscribe' to certain jobs, that way we do not have partners constantly hitting up the system looking for jobs to perform (which could act like a DDoS attack from all the incoming request)*

```Python
def handle_validator_change_state(self, packet: bytes) -> None:

        '''
        Handles validator change state packet
        '''

        logger.info("Handling Validator Change State")

        new_state = packet[2:].decode("utf-8")

        logger.info(f"Validator changed to state: {new_state}")

        ...
```

This handler is used when a validator is changing its state in the system (as you want to advertise that on the network). Examples of a change state is when a validator goes from being an active validator to a passive validator or when a validator is pending (getting ready to become an active validator). This should only go out to the active validators as I don't see a reason to advertise this to all validators *but I could be wrong*

Current states:

- Discovery
- Sync
- Pending
- Redirect
- Active
- Error

```Python
def handle_report_packet(self, packet_data: bytearray) -> None:

        '''
        Handles the report packet, extracting the necessary information

        and logging or acting on the report.
        '''

        reporter = PacketUtils._decode_public_key(packet_data[:64])

        reported = PacketUtils._decode_public_key(packet_data[64:128])

        reason = PacketUtils._decode_string(packet_data[128:])

        logger.info(f"Received report from {reporter} about {reported} for reason: {reason}.")
```

This is meant to handle any reports that come in regarding a user who acts inappropriately. This system will be used to update perception scores and will be logged on the network as issues occur. 

## Example Offenses (Validator-Specific)

Although this handler applies to **all user types**, some examples of validator-specific reports might include:

- Failing to reach consensus with peers repeatedly (proof: mismatched job hashes or late votes)
- Excessive latency during job routing
- Ignoring or rejecting valid job requests
- Not responding to time sync or shut-up packets
- Failing to submit receipts after job assignment

*NOTE: This handler will be true of all user types. There are more negative behaviors that are listed inside of the readme document*

The payload is parsed as follows:

- `reporter`: Public key of the reporting user (first 64 bytes)
- `reported`: Public key of the user being reported (next 64 bytes)
- `reason`: UTF-8 string describing the offense (remainder of payload)

*THOUGHT: We should build a report registry that lists all the possible report types and then store them as ENUMs so that we can just enter a code rather than a text description (space saving)*

```Python
def handle_perception_update_packet(self, packet_data: bytearray) -> None:

        '''
        Handles the perception score update packet, updating the perception score

        for the user in the local validator's perception score table.
        '''

        user_id = PacketUtils._decode_public_key(packet_data[:64])

        new_score = int.from_bytes(packet_data[64:68], byteorder='big')

  

        logger.info(f"Updating perception score for user {user_id} to {new_score}.")
```

This is one of UndChain's main core features; this will define how perception scores will be updated on the network. It will 

1. Perform a consensus with the active validators using things such as a **report packet** in the event that a user is behaving negatively or a **achievement packet** for when a user performs a good service for the network. 
2. Once validators preform consensus they share the update with the partners who then log it on network storage. Network storage is where we will bootstrap the perception score of each user (along with other items such as account age and a history of their score); this is done since you do NOT want validators holding this information for every user on the network. It also provides balance so that validators are not the only holders of this information, partners will also need to agree. 

*NOTE: Users can choose to block validators or partners, but this can NEVER happen the other way on the main chain. I do see some chains blocking users due to malicious behavior but that should be via perception score. Don't want to block on main chain as that would prevent users from transferring funds. While a block will NOT produce a negative perception score, a partner or validator who is consistently blocked will decrease in perception score.* 


```Python
    def get_packet_type(self, packet: bytes) -> PacketType:

        '''
        Extracts the packet type from the first two bytes of the packet.
        '''

        try:

            pack_type_value = struct.unpack(">H", packet[:2])[0] # First two bytes represent the packet type

            return PacketType(pack_type_value)

        except Exception as e:

            logger.error(f'Failed to extract packet type: {e}')

            raise ValueError(f'Unknown packet type from packet {e}')
```

This handlers only job is to classify the packet type that is coming in so that it can be processed correctly (this is a helper function). It looks at the first two bytes in order to determine the packet type.

```Python
# Example use

if __name__ == "__main__":

    packet_generator = PacketGenerator("2024.10.09.1")

    handler = PacketHandler(packet_generator)
  
    # Simulate generating a VALIDATOR_REQUEST packet using PacketGenerator

    public_key = b"validator_pub_key_12345"

    sample_packet: bytes = packet_generator.generate_validator_request(public_key)

    print(f'Sample Packet: {sample_packet}')

    # Now pass the sample packet to the PacketHandler for processing

    return_packet = handler.handle_packet(sample_packet)

    print(f"Generated return packet: {return_packet}")

    if return_packet:

        return_packet = handler.handle_packet(return_packet)

        print(f"Interpreted confirmation packet: {return_packet}")
```

This `__main__` block serves as a **self-test and usage demo** for the `PacketHandler` system.

### Step-by-Step:

1. A `PacketGenerator` is initialized with a version number (e.g. `2024.10.09.1`) — this will define how the packet is structured.
2. A fake public key is generated to simulate a new validator joining the network.
3. A **VALIDATOR_REQUEST** packet is created and passed to the handler.
4. The handler decodes it and returns a **VALIDATOR_CONFIRMATION** response.
5. That confirmation is then re-processed to demonstrate two-way packet flow.
    

This allows you to quickly validate that:

- Packets are **being encoded correctly**
- The correct handler method is being invoked
- The payload is being decoded accurately

This example MUST be updated as we add more methods to this class so that we can ensure its being tested correctly.


## Summary

The `PacketHandler` class centralizes **packet classification and response logic** for all known packet types on the network.

- It separates concerns between **packet generation** and **packet interpretation**
- It routes each incoming packet to its appropriate handler method via a clean dictionary map
- It logs and decodes payloads in a consistent, predictable format
- It will evolve as new user types and services are added to UndChain (e.g., partners, clients, chain owners)
    

> Right now, the focus is on validator packet types — but this architecture ensures that adding new user-specific packet flows (like partner-to-client job confirmations or perception updates) will be easy to expand without cluttering the validator logic.

---

# Packet Generator

The `PacketGenerator` class is the **counterpart to the `PacketHandler`** — where the handler **listens**, the generator **speaks**. This class creates properly formatted packets that can be transmitted across the UndChain network.

It serves as a **centralized factory** for all packet types a validator may need to send — whether it's:

- A response to an incoming request (e.g., validator confirmation)
- A status announcement (e.g., validator change state)
- A broadcast (e.g., job file or perception update)
    

Standardizing packets through this generator:

- Keeps your message format consistent
- Makes future protocol upgrades easier to manage
- Keeps packet versioning and encoding in a single, testable place

```Python
class PacketType(Enum):

    VALIDATOR_REQUEST = 1
    VALIDATOR_CONFIRMATION = 2
    VALIDATOR_STATE = 3
    VALIDATOR_LIST_REQUEST = 4
    VALIDATOR_LIST_RESPONSE = 5
    JOB_FILE = 6
    PAYOUT_FILE = 7
    SHUT_UP = 8
    LATENCY = 9
    CONVERGENCE = 10
    SYNC_CO_CHAIN = 11
    SHARE_RULES = 12
    JOB_REQUEST = 13
    VALIDATOR_CHANGE_STATE = 14
    VALIDATOR_VOTE = 15
    RETURN_ADDRESS = 16
    REPORT = 17
    PERCEPTION_UPDATE = 18
```

The `PacketType` ENUM defines all currently supported packet categories. Instead of using plain-text labels (which are longer and inconsistent), we assign each a fixed **integer value**, which improves:

- **Network efficiency** (fewer bytes)
- **Parsing speed** (known header size)
- **Protocol upgradeability** (easy to insert new packet types without changing the parser logic)

This ENUM is used at the beginning of every packet header to help the handler identify which operation to execute.