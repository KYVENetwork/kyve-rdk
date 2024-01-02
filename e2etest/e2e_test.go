package e2etest

import (
	"context"
	"cosmossdk.io/math"
	"fmt"
	querytypes "github.com/KYVENetwork/chain/x/query/types"
	"github.com/KYVENetwork/kyvejs/e2etest/utils"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkquerytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/docker/docker/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestIntegrations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, fmt.Sprint("Kyvejs integration tests"))
}

func getFinalizedBundles(client querytypes.QueryBundlesClient) *querytypes.QueryFinalizedBundlesResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	params := &querytypes.QueryFinalizedBundlesRequest{
		PoolId:     0,
		Pagination: &sdkquerytypes.PageRequest{},
	}

	res, err := client.FinalizedBundlesQuery(ctx, params)
	Expect(err).To(BeNil())

	return res
}

var _ = Describe(fmt.Sprintf("e2e Tests"), Ordered, func() {
	var kyveChain *cosmos.CosmosChain

	var ctx context.Context
	var client *client.Client
	var network string
	var interchain *interchaintest.Interchain
	var broadcaster *cosmos.Broadcaster
	var queryBundlesClient querytypes.QueryBundlesClient

	// TODO: remove this var
	var kyveWallet *cosmos.CosmosWallet

	var protocolRunner *utils.ProtocolRunner

	BeforeAll(func() {
		numFullNodes := 0
		numValidators := 2
		factory := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(GinkgoT()), []*interchaintest.ChainSpec{
			{
				Name:          "kyve",
				ChainConfig:   utils.MainnetConfig,
				NumValidators: &numValidators,
				NumFullNodes:  &numFullNodes,
			},
		})

		chains, err := factory.Chains(GinkgoT().Name())
		Expect(err).To(BeNil())
		kyveChain = chains[0].(*cosmos.CosmosChain)

		interchain = interchaintest.NewInterchain().AddChain(kyveChain)

		ctx = context.Background()
		client, network = interchaintest.DockerSetup(GinkgoT())

		protocolRunner = utils.NewProtocolRunner()
		err = protocolRunner.Build()
		Expect(err).To(BeNil())

		err = interchain.Build(ctx, nil, interchaintest.InterchainBuildOptions{
			TestName:         GinkgoT().Name(),
			Client:           client,
			NetworkID:        network,
			SkipPathCreation: true,
		})
		Expect(err).To(BeNil())

		broadcaster = cosmos.NewBroadcaster(GinkgoT(), kyveChain)
		broadcaster.ConfigureClientContextOptions(func(clientContext sdkclient.Context) sdkclient.Context {
			return clientContext.WithBroadcastMode(flags.BroadcastSync)
		})

		queryBundlesClient = querytypes.NewQueryBundlesClient(sdkclient.Context{
			Client: kyveChain.GetNode().Client,
		})

		//wallets := interchaintest.GetAndFundTestUsers(
		//	GinkgoT(), ctx, "testuser", math.NewInt(10_000_000_000), kyveChain,
		//)
		//kyveWallet = wallets[0].(*cosmos.CosmosWallet)

		for _, mnemonic := range utils.Mnemonics {
			_, err = interchaintest.GetAndFundTestUserWithMnemonic(ctx, "e2e-test", mnemonic, math.NewInt(10_000_000_000_000), kyveChain)
			Expect(err).To(BeNil())
		}

		err = protocolRunner.Run(client, network, kyveChain.GetAPIAddress(), kyveChain.GetRPCAddress())
		Expect(err).To(BeNil())
	})

	AfterAll(func() {
		err := interchain.Close()
		if err != nil {
			fmt.Println(err)
		}

		err = protocolRunner.Cleanup(client)
		if err != nil {
			fmt.Println(err)
		}
	})

	It("Test finalized bundles", func() {
		// Wait for 4 finalized bundles to be created
		waitForBundles := 4
		err := testutil.WaitForCondition(10*time.Minute, 5*time.Second, func() (bool, error) {
			bundles := getFinalizedBundles(queryBundlesClient)
			return len(bundles.FinalizedBundles) == waitForBundles, nil
		})
		if err != nil {
			// If the test times out, print the finalized bundles
			fmt.Println(getFinalizedBundles(queryBundlesClient))
			Fail(err.Error())
		}

		// TODO: remove this check
		bundles := getFinalizedBundles(queryBundlesClient)
		Expect(len(bundles.FinalizedBundles)).To(Equal(waitForBundles))

		// TODO: remove this check
		Expect(kyveWallet).NotTo(BeNil())
	})
})
