// Code generated by "genpubsub"; DO NOT EDIT.

package pubsub

// Condition is the underlying cause of a pubsub error.
type Condition uint32

// Valid pubsub Conditions.
const (
	ClosedNode             Condition = iota // closed-node
	ConfigRequired                          // configuration-required
	InvalidJID                              // invalid-jid
	InvalidOptions                          // invalid-options
	InvalidPayload                          // invalid-payload
	InvalidSubID                            // invalid-subid
	ItemForbidden                           // item-forbidden
	ItemRequired                            // item-required
	JIDRequired                             // jid-required
	MaxItemsExceeded                        // max-items-exceeded
	MaxNodesExceeded                        // max-nodes-exceeded
	NodeIDRequired                          // nodeid-required
	NotInRosterGroup                        // not-in-roster-group
	NotSubscribed                           // not-subscribed
	PayloadTooBig                           // payload-too-big
	PayloadRequired                         // payload-required
	PendingSubscription                     // pending-subscription
	PresenceRequired                        // presence-subscription-required
	SubIDRequired                           // subid-required
	TooManySubscriptions                    // too-many-subscriptions
	Unsupported                             // unsupported
	UnsupportedAccessModel                  // unsupported-access-model
)

// Feature is a specific pubsub feature that may be reported in an error as
// being unsupported.
type Feature uint32

// Valid pubsub Features.
const (
	FeatureAccessAuthorize           Feature = iota // access-authorize
	FeatureAccessOpen                               // access-open
	FeatureAccessPresence                           // access-presence
	FeatureAccessRoster                             // access-roster
	FeatureAccessWhitelist                          // access-whitelist
	FeatureAutoCreate                               // auto-create
	FeatureAutoSubscribe                            // auto-subscribe
	FeatureCollections                              // collections
	FeatureConfigNode                               // config-node
	FeatureCreateAndConfigure                       // create-and-configure
	FeatureCreateNodes                              // create-nodes
	FeatureDeleteItems                              // delete-items
	FeatureDeleteNodes                              // delete-nodes
	FeatureFilteredNotifications                    // filtered-notifications
	FeatureGetPending                               // get-pending
	FeatureInstantNodes                             // instant-nodes
	FeatureItemIDs                                  // item-ids
	FeatureLastPublished                            // last-published
	FeatureLeasedSubscription                       // leased-subscription
	FeatureManageSubscriptions                      // manage-subscriptions
	FeatureMemberAffiliation                        // member-affiliation
	FeatureMetaData                                 // meta-data
	FeatureModifyAffiliations                       // modify-affiliations
	FeatureMultiCollection                          // multi-collection
	FeatureMultiSubscribe                           // multi-subscribe
	FeatureOutcastAffiliation                       // outcast-affiliation
	FeaturePersistentItems                          // persistent-items
	FeaturePresenceNotifications                    // presence-notifications
	FeaturePresenceSubscribe                        // presence-subscribe
	FeaturePublish                                  // publish
	FeaturePublishOptions                           // publish-options
	FeaturePublishOnlyAffiliation                   // publish-only-affiliation
	FeaturePublisherAffiliation                     // publisher-affiliation
	FeaturePurgeNodes                               // purge-nodes
	FeatureRetractItems                             // retract-items
	FeatureRetrieveAffiliations                     // retrieve-affiliations
	FeatureRetrieveDefault                          // retrieve-default
	FeatureRetrieveItems                            // retrieve-items
	FeatureRetrieveSubscriptions                    // retrieve-subscriptions
	FeatureSubscribe                                // subscribe
	FeatureSubscriptionOptions                      // subscription-options
	FeatureSubscriptionNotifications                // subscription-notifications
)

// SubType represents the state of a particular subscription.
type SubType uint8

// A list of possible subscription types.
const (
	SubNone         SubType = iota // none
	SubPending                     // pending
	SubSubscribed                  // subscribed
	SubUnconfigured                // unconfigured
)
