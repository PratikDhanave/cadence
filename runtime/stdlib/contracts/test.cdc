/// Test contract is the standard library that provides testing functionality in Cadence.
///
pub contract Test {

    /// Blockchain emulates a real network.
    ///
    pub struct Blockchain {

        pub let backend: AnyStruct{BlockchainBackend}

        init(backend: AnyStruct{BlockchainBackend}) {
            self.backend = backend
        }

        /// Executes a script and returns the script return value and the status.
        /// `returnValue` field of the result will be `nil` if the script failed.
        ///
        pub fun executeScript(_ script: String, _ arguments: [AnyStruct]): ScriptResult {
            return self.backend.executeScript(script, arguments)
        }

        /// Creates a signer account by submitting an account creation transaction.
        /// The transaction is paid by the service account.
        /// The returned account can be used to sign and authorize transactions.
        ///
        pub fun createAccount(): Account {
            return self.backend.createAccount()
        }

        /// Add a transaction to the current block.
        ///
        pub fun addTransaction(_ tx: Transaction) {
            self.backend.addTransaction(tx)
        }

        /// Executes the next transaction in the block, if any.
        /// Returns the result of the transaction, or nil if no transaction was scheduled.
        ///
        pub fun executeNextTransaction(): TransactionResult? {
            return self.backend.executeNextTransaction()
        }

        /// Commit the current block.
        /// Committing will fail if there are un-executed transactions in the block.
        ///
        pub fun commitBlock() {
            self.backend.commitBlock()
        }

        /// Executes a given transaction and commit the current block.
        ///
        pub fun executeTransaction(_ tx: Transaction): TransactionResult {
            self.addTransaction(tx)
            let txResult = self.executeNextTransaction()!
            self.commitBlock()
            return txResult
        }

        /// Executes a given set of transactions and commit the current block.
        ///
        pub fun executeTransactions(_ transactions: [Transaction]): [TransactionResult] {
            for tx in transactions {
                self.addTransaction(tx)
            }

            var results: [TransactionResult] = []
            for tx in transactions {
                let txResult = self.executeNextTransaction()!
                results.append(txResult)
            }

            self.commitBlock()
            return results
        }

        /// Deploys a given contract, and initilizes it with the arguments.
        ///
        pub fun deployContract(
            name: String,
            code: String,
            account: Account,
            arguments: [AnyStruct]
        ): Error? {
            return self.backend.deployContract(
                name: name,
                code: code,
                account: account,
                arguments: arguments
            )
        }

        /// Set the configuration to be used by the blockchain.
        /// Overrides any existing configuration.
        ///
        pub fun useConfiguration(_ configuration: Configuration) {
            self.backend.useConfiguration(configuration)
        }

        /// Returns all the logs from the blockchain, up to the calling point.
        ///
        pub fun logs(): [String] {
            return self.backend.logs()
        }

        /// Returns the service account of the blockchain. Can be used to sign
        /// transactions with this account.
        ///
        pub fun serviceAccount(): Account {
            return self.backend.serviceAccount()
        }

        /// Returns all events emitted from the blockchain.
        ///
        pub fun events(): [AnyStruct] {
            return self.backend.events(nil)
        }

        /// Returns all events emitted from the blockchain,
        /// filtered by type.
        ///
        pub fun eventsOfType(_ type: Type): [AnyStruct] {
            return self.backend.events(type)
        }

        /// Resets the state of the blockchain to the given height.
        ///
        pub fun reset(to height: UInt64) {
            self.backend.reset(to: height)
        }
    }

    pub struct Matcher {

        pub let test: ((AnyStruct): Bool)

        pub init(test: ((AnyStruct): Bool)) {
            self.test = test
        }

        /// Combine this matcher with the given matcher.
        /// Returns a new matcher that succeeds if this and the given matcher succeed.
        ///
        pub fun and(_ other: Matcher): Matcher {
            return Matcher(test: fun (value: AnyStruct): Bool {
                return self.test(value) && other.test(value)
            })
        }

        /// Combine this matcher with the given matcher.
        /// Returns a new matcher that succeeds if this or the given matcher succeed.
        /// If this matcher succeeds, then the other matcher would not be tested.
        ///
        pub fun or(_ other: Matcher): Matcher {
            return Matcher(test: fun (value: AnyStruct): Bool {
                return self.test(value) || other.test(value)
            })
        }
    }

    /// ResultStatus indicates status of a transaction or script execution.
    ///
    pub enum ResultStatus: UInt8 {
        pub case succeeded
        pub case failed
    }

    /// Result is the interface to be implemented by the various execution
    /// operations, such as transactions and scripts.
    ///
    pub struct interface Result {
        /// The result status of an executed operation.
        ///
        pub let status: ResultStatus

        /// The optional error of an executed operation.
        ///
        pub let error: Error?
    }

    /// The result of a transaction execution.
    ///
    pub struct TransactionResult: Result {
        pub let status: ResultStatus
        pub let error: Error?

        init(status: ResultStatus, error: Error?) {
            self.status = status
            self.error = error
        }
    }

    /// The result of a script execution.
    ///
    pub struct ScriptResult: Result {
        pub let status: ResultStatus
        pub let returnValue: AnyStruct?
        pub let error: Error?

        init(status: ResultStatus, returnValue: AnyStruct?, error: Error?) {
            self.status = status
            self.returnValue = returnValue
            self.error = error
        }
    }

    // Error is returned if something has gone wrong.
    //
    pub struct Error {
        pub let message: String

        init(_ message: String) {
            self.message = message
        }
    }

    /// Account represents info about the account created on the blockchain.
    ///
    pub struct Account {
        pub let address: Address
        pub let publicKey: PublicKey

        init(address: Address, publicKey: PublicKey) {
            self.address = address
            self.publicKey = publicKey
        }
    }

    /// Configuration to be used by the blockchain.
    /// Can be used to set the address mappings.
    ///
    pub struct Configuration {
        pub let addresses: {String: Address}

        init(addresses: {String: Address}) {
            self.addresses = addresses
        }
    }

    /// Transaction that can be submitted and executed on the blockchain.
    ///
    pub struct Transaction {
        pub let code: String
        pub let authorizers: [Address]
        pub let signers: [Account]
        pub let arguments: [AnyStruct]

        init(code: String, authorizers: [Address], signers: [Account], arguments: [AnyStruct]) {
            self.code = code
            self.authorizers = authorizers
            self.signers = signers
            self.arguments = arguments
        }
    }

    /// BlockchainBackend is the interface to be implemented by the backend providers.
    ///
    pub struct interface BlockchainBackend {

        /// Executes a script and returns the script return value and the status.
        /// `returnValue` field of the result will be `nil` if the script failed.
        ///
        pub fun executeScript(_ script: String, _ arguments: [AnyStruct]): ScriptResult

        /// Creates a signer account by submitting an account creation transaction.
        /// The transaction is paid by the service account.
        /// The returned account can be used to sign and authorize transactions.
        ///
        pub fun createAccount(): Account

        /// Add a transaction to the current block.
        ///
        pub fun addTransaction(_ tx: Transaction)

        /// Executes the next transaction in the block, if any.
        /// Returns the result of the transaction, or nil if no transaction was scheduled.
        ///
        pub fun executeNextTransaction(): TransactionResult?

        /// Commit the current block.
        /// Committing will fail if there are un-executed transactions in the block.
        ///
        pub fun commitBlock()

        /// Deploys a given contract, and initilizes it with the arguments.
        ///
        pub fun deployContract(
            name: String,
            code: String,
            account: Account,
            arguments: [AnyStruct]
        ): Error?

        /// Set the configuration to be used by the blockchain.
        /// Overrides any existing configuration.
        ///
        pub fun useConfiguration(_ configuration: Configuration)

        /// Returns all the logs from the blockchain, up to the calling point.
        ///
        pub fun logs(): [String]

        /// Returns the service account of the blockchain. Can be used to sign
        /// transactions with this account.
        ///
        pub fun serviceAccount(): Account

        /// Returns all events emitted from the blockchain, optionally filtered
        /// by type.
        ///
        pub fun events(_ type: Type?): [AnyStruct]

        /// Resets the state of the blockchain to the given height.
        ///
        pub fun reset(to height: UInt64)
    }

    /// Returns a new matcher that negates the test of the given matcher.
    ///
    pub fun not(_ matcher: Matcher): Matcher {
        return Matcher(test: fun (value: AnyStruct): Bool {
            return !matcher.test(value)
        })
    }

    /// Returns a new matcher that checks if the given test value is either
    /// a ScriptResult or TransactionResult and the ResultStatus is succeeded.
    /// Returns false in any other case.
    ///
    pub fun beSucceeded(): Matcher {
        return Matcher(test: fun (value: AnyStruct): Bool {
            return (value as! {Result}).status == ResultStatus.succeeded
        })
    }

    /// Returns a new matcher that checks if the given test value is either
    /// a ScriptResult or TransactionResult and the ResultStatus is failed.
    /// Returns false in any other case.
    ///
    pub fun beFailed(): Matcher {
        return Matcher(test: fun (value: AnyStruct): Bool {
            return (value as! {Result}).status == ResultStatus.failed
        })
    }

    /// Returns a new matcher that checks if the given test value is nil.
    ///
    pub fun beNil(): Matcher {
        return Matcher(test: fun (value: AnyStruct): Bool {
            return value == nil
        })
    }

    /// Asserts that the result status of an executed operation, such as
    /// a script or transaction, has failed and contains the given error
    /// message.
    ///
    pub fun assertError(_ result: {Result}, errorMessage: String) {
        pre {
            result.status == ResultStatus.failed: "no error was found"
        }

        var found = false
        let msg = result.error!.message
        let msgLength = msg.length - errorMessage.length + 1
        var i = 0
        while i < msgLength {
            if msg.slice(from: i, upTo: i + errorMessage.length) == errorMessage {
                found = true
                break
            }
            i = i + 1
        }

        assert(found, message: "the error message did not contain the given sub-string")
    }

}
