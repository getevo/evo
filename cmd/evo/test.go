package main

import (
	"gorm.io/gorm"
	"time"
)

type App struct{}

func (App)SomeFunction()  {

}

// Feature list of product features description
type Feature struct{
	IDFeature          	int                	`gorm:"primary_key;column:id_feature;type:int(10) unsigned;not null" json:"id_feature"` // product feature numeric identifier (autoincrement)
	IDProduct          	int                	`gorm:"index:id_product;column:id_product;type:int(10) unsigned;not null" json:"-"`
	BadgeFeature       	string             	`gorm:"column:badge_feature;type:varchar(100);not null" json:"badge_feature"`
	DescrFeature       	string             	`gorm:"column:descr_feature;type:varchar(255);not null" json:"descr_feature"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"-"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"-"`
	DeletedAt          	*time.Time         	`gorm:"column:deleted_at;type:datetime" json:"-"`
}

func (Feature)TableName() string {
	return "prd_feature"
}
// Packet packet
type Packet struct{
	IDPacket           	int                	`gorm:"primary_key;column:id_packet;type:int(10) unsigned;not null" json:"-"`   // packet numeric identifier (autoincrement)
	NamePacket         	string             	`gorm:"unique;column:name_packet;type:varchar(64);not null" json:"name_packet"` // packet component name (extended name/description)
	DescPacket         	string             	`gorm:"column:desc_packet;type:varchar(255);not null" json:"desc_packet"`       // packet component description, contains information about what the user with this role can do
	Type               	string             	`gorm:"column:type;type:enum('subscription','ppv')" json:"type"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"-"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"-"`
	DeletedAt          	time.Time          	`gorm:"column:deleted_at;type:datetime" json:"deleted_at"`
}
func (Packet)TableName() string {
	return "prd_packet"
}

// PacketComponent relation between packet and component
type PacketComponent struct{
	IDPacketComponent  	int                	`gorm:"primary_key;column:id_packet_component;type:int(10) unsigned;not null" json:"-"`                                                // packet/component numeric identifier
	IDPacket           	int                	`gorm:"unique_index:id_packet;index:fk_prd_packet_component_product;column:id_packet;type:int(10) unsigned;not null" json:"id_packet"` // packet numeric identifier
	PrdPacket          	Packet             	`gorm:"association_foreignkey:id_packet;foreignkey:id_packet" json:"prd_packet_list"`                                                  // packet
	IDComponent        	int                	`gorm:"unique_index:id_packet;index:fk_prd_packet_component_component;column:id_component;type:int(10);not null" json:"id_component"`
	PrdComponent       	Component          	`gorm:"association_foreignkey:id_component;foreignkey:id_component" json:"prd_component_list"` // product component
	NumComponent       	int                	`gorm:"column:num_component;type:int(10)" json:"num_component"`
	DtFrom             	time.Time          	`gorm:"column:dt_from;type:datetime" json:"dt_from"`
	DtTo               	time.Time          	`gorm:"column:dt_to;type:datetime" json:"dt_to"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"-"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"-"`
	DeletedAt          	time.Time          	`gorm:"column:deleted_at;type:datetime" json:"-"`
}

func (PacketComponent)TableName() string {
	return "prd_packet_component"
}


// Product product label to assign price based on sales channel (country)
type Product struct{
	IDProduct          	int                	`gorm:"primary_key;column:id_product;type:int(10) unsigned;not null" json:"id_product"`                                                         // product numeric identifier (autoincrement)
	CodeProduct        	string             	`gorm:"column:code_product;type:varchar(32)" json:"code_product" form:"code_product" validate:"format=alnum required"`                                                                      // product code (short name)
	NameProduct        	string             	`gorm:"column:name_product;type:varchar(64)" json:"name_product" form:"name_product" validate:"format=strict_html required"`                                                                      // product name
	DescProduct        	string             	`gorm:"column:desc_product;type:varchar(64)" json:"desc_product" form:"desc_product" validate:"format=strict_html"`                                                                      // product description
	IDPacket           	int                	`gorm:"unique_index:uk_prd_product;column:id_packet;type:int(10) unsigned;not null" json:"id_packet" validate:"format=number"  form:"id_packet"`                                  // packet numeric identifier
	PrdPacket          	Packet             	`gorm:"association_foreignkey:id_packet;foreignkey:id_packet" json:"prd_packet_list"`                                                  // packet
	Features           	[]Feature          	`gorm:"association_foreignkey:IDProduct;foreignkey:IDProduct" json:"features"`
	TrialPeriodDays    	int                	`gorm:"column:trial_period_days;type:int(10) unsigned" json:"trial_period_days" form:"trial_period_days" validate:"format=number"`
	PaymentFrequency   	string             	`gorm:"column:payment_frequency;type:enum('W','M','Y','1S');not null" json:"payment_frequency" form:"payment_frequency" validate:"one_of=W,M,Y,1S"`
	PaymentInterval    	uint8              	`gorm:"column:payment_interval;type:tinyint(3) unsigned" json:"payment_interval" form:"payment_interval" validate:""`            // number of period to pay, for example with payment_frequency = M num_period may be 10
	PricePeriod        	float64            	`gorm:"column:price_period;type:decimal(12,2)" json:"price_period" form:"price_period"`              // period gross price
	PriceExclTax       	float64            	`gorm:"column:price_excl_tax;type:decimal(12,2);not null" json:"price_excl_tax" form:"price_excl_tax"` // net price
	PriceRetail        	float64            	`gorm:"column:price_retail;type:decimal(12,2);not null" json:"price_retail" form:"price_retail"`     // gross price
	Currency           	string             	`gorm:"column:currency;type:char(3);not null" json:"currency" form:"currency"`                   // currency code (short name)
	IDRule             	int                	`gorm:"index:fk_prd_product_rule;column:id_rule;type:int(10) unsigned;not null" json:"id_rule" form:"id_rule"`
	ValidityFrom       	time.Time          	`gorm:"column:validity_from;type:datetime" json:"validity_from" form:"validity_from"` // start validity of the product, before this date the product in not available for the sell
	ValidityTo         	time.Time          	`gorm:"column:validity_to;type:datetime" json:"validity_to" form:"validity_to"`     // end validity of the product, after this date the product in not available for the sell
	VisibleFrom        	time.Time          	`gorm:"column:visible_from;type:datetime" json:"visible_from" form:"visible_from"`
	VisibleTo          	time.Time          	`gorm:"column:visible_to;type:datetime" json:"visible_to" form:"visible_to"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"-"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"-"`
	DeletedAt          	*time.Time         	`gorm:"column:deleted_at;type:datetime" json:"-"`
}
func (Product)TableName() string {
	return "prd_product"
}


// Component product component
type Component struct{
	IDComponent        	int                	`gorm:"primary_key;column:id_component;type:int(10);not null" json:"-"`                              // component id
	CodeComponent      	string             	`gorm:"unique;column:code_component;type:varchar(32);not null" json:"code_component"`                // component code (short name)
	NameComponent      	string             	`gorm:"unique;column:name_component;type:varchar(64);not null" json:"name_component"`                // component name (extended name/description); the product component name may be not unique
	DescComponent      	string             	`gorm:"column:desc_component;type:varchar(255);not null" json:"desc_component"`                      // component description, contains information about what the user with this role can do
	IDResource         	int                	`gorm:"index:idx_id_resource;column:id_resource;type:int(10) unsigned" json:"id_resource"`           // starting id in the table specified
	ResourceType       	string             	`gorm:"index:idx_resource_type;column:resource_type;type:varchar(30);not null" json:"resource_type"` // table specified
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"created_at"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"updated_at"`
	DeletedAt          	time.Time          	`gorm:"column:deleted_at;type:datetime" json:"deleted_at"`
}
func (Component)TableName() string {
	return "prd_component"
}


// Order [...]
type Order struct{
	IDOrder            	int                	`gorm:"primary_key;column:id_order;type:int(10) unsigned;not null" json:"id_order"`
	ParentIDOrder      	*int               	`gorm:"index:fk_parent_id_order;column:parent_id_order;type:int(10) unsigned" json:"parent_id_order"`
	Order              	*Order             	`gorm:"association_foreignkey:parent_id_order;foreignkey:id_order" json:"order_list"`
	IDPgJob            	int                	`gorm:"column:id_pg_job;type:int(11);not null" json:"id_pg_job"`
	NumItemOrder       	int64              	`gorm:"column:num_item_order;type:int(11) unsigned" json:"num_item_order"`
	DateOrder          	time.Time          	`gorm:"column:date_order;type:datetime;not null" json:"date_order"`
	Status             	string             	`gorm:"column:status;type:enum('pending','failed','completed','trial','canceled')" json:"status"`
	Cancellation       	*time.Time         	`gorm:"column:cancellation;type:datetime" json:"cancellation"`
	NoteOrder          	string             	`gorm:"column:note_order;type:varchar(255);not null" json:"note_order"`
	IDProduct          	int                	`gorm:"index:fk_id_product_order;column:id_product;type:int(10) unsigned" json:"id_product"`
	PrdProduct         	Product            	`gorm:"association_foreignkey:id_product;foreignkey:id_product" json:"prd_product_list"` // product label to assign price based on sales channel (country)
	Quantity           	int                	`gorm:"column:quantity;type:int(10) unsigned;not null" json:"quantity"`
	TrialPeriodDays    	int                	`gorm:"column:trial_period_days;type:int(10) unsigned" json:"trial_period_days"`
	Currency           	string             	`gorm:"column:currency;type:char(3);not null" json:"currency"`
	TaxCode            	string             	`gorm:"column:tax_code;type:varchar(5);not null" json:"tax_code"`
	UnitRetailPrice    	float64            	`gorm:"column:unit_retail_price;type:decimal(12,2);not null" json:"unit_retail_price"`
	UnitNetPrice       	float64            	`gorm:"column:unit_net_price;type:decimal(12,2);not null" json:"unit_net_price"`
	UnitTaxAmount      	float64            	`gorm:"column:unit_tax_amount;type:decimal(12,2);not null" json:"unit_tax_amount"`
	RetailPrice        	float64            	`gorm:"column:retail_price;type:decimal(12,2);not null" json:"retail_price"`
	NetPrice           	float64            	`gorm:"column:net_price;type:decimal(12,2);not null" json:"net_price"`
	TaxAmount          	float64            	`gorm:"column:tax_amount;type:decimal(12,2);not null" json:"tax_amount"`
	DiscountAmount     	float64            	`gorm:"column:discount_amount;type:decimal(12,2)" json:"discount_amount"`
	DiscountPercent    	int                	`gorm:"column:discount_percent;type:int(11) unsigned" json:"discount_percent"`
	GiftCardAmount     	float64            	`gorm:"column:gift_card_amount;type:decimal(12,2)" json:"gift_card_amount"`
	IDCoupon           	*int               	`gorm:"index:fk_id_coupon_order;column:id_coupon;type:int(10) unsigned" json:"id_coupon"`
	Coupon             	Coupon             	`gorm:"association_foreignkey:id_coupon;foreignkey:id" json:"coupon_list"`
	IDUser             	uint               	`gorm:"index:order_id_user;column:id_user;type:int(10) unsigned" json:"id_user"`
	IDBillingAddress   	*int               	`gorm:"index:order_id_billing_address;column:id_billing_address;type:int(10) unsigned" json:"id_billing_address"`
	OrderBillingaddress	OrderBillingaddress	`gorm:"association_foreignkey:id_billing_address;foreignkey:id_billing_address" json:"order_billingaddress_list"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"created_at"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"updated_at"`
	DeletedAt          	*time.Time         	`gorm:"column:deleted_at;type:datetime" json:"deleted_at"`
}

func (Order)TableName() string {
	return "order"
}



// OrderBillingaddress [...]
type OrderBillingaddress struct{
	IDBillingAddress   	int                	`gorm:"primary_key;column:id_billing_address;type:int(10) unsigned;not null" json:"-"`
	IDUser             	int                	`gorm:"index:billingaddress_id_user;column:id_user;type:int(10) unsigned" json:"id_user"`
	Title              	string             	`gorm:"column:title;type:varchar(64);not null" json:"title"`
	FirstName          	string             	`gorm:"column:first_name;type:varchar(255);not null" json:"first_name"`
	LastName           	string             	`gorm:"column:last_name;type:varchar(255);not null" json:"last_name"`
	Address            	string             	`gorm:"column:address;type:varchar(255);not null" json:"address"`
	Postcode           	string             	`gorm:"column:postcode;type:varchar(64);not null" json:"postcode"`
	IsoCountry         	string             	`gorm:"index:billingaddress_iso_country;column:iso_country;type:varchar(2)" json:"iso_country"`
	//Country          	Country            	`gorm:"association_foreignkey:iso_country;foreignkey:iso_country" json:"country_list"`
	SearchText         	string             	`gorm:"column:search_text;type:longtext;not null" json:"search_text"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"created_at"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"updated_at"`
	DeletedAt          	time.Time          	`gorm:"column:deleted_at;type:datetime" json:"deleted_at"`
}

// UserCoupon [...]
type UserCoupon struct{
gorm.Model
	IDUser             	int                	`gorm:"index:fk_user_coupon_idx;column:id_user;type:int(10) unsigned;not null" json:"id_user"`
	IDCoupon           	int                	`gorm:"index:fk_coupon_idx;column:id_coupon;type:int(10) unsigned;not null" json:"id_coupon"`
	Coupon             	Coupon             	`gorm:"association_foreignkey:id_coupon;foreignkey:id" json:"coupon_list"`
	Status             	string             	`gorm:"column:status;type:enum('available','pending','applied');not null" json:"status"`
	ExpireAt           	time.Time          	`gorm:"column:expire_at;type:datetime" json:"expire_at"`
}


// Coupon [...]
type Coupon struct{
gorm.Model
	IDPaymentMethod    	int                	`gorm:"column:id_payment_method;type:int(11)" json:"id_payment_method"`
	Name               	string             	`gorm:"column:name;type:char(55);not null" json:"name"`
	Key                	string             	`gorm:"unique;column:key;type:char(55);not null" json:"key"`
	PercentOff         	float64            	`gorm:"column:percent_off;type:decimal(6,3) unsigned" json:"percent_off"`
	Discount           	int                	`gorm:"column:discount;type:int(3) unsigned" json:"discount"`
	Currency           	string             	`gorm:"column:currency;type:char(3);not null" json:"currency"`
	KeyStripe          	string             	`gorm:"column:key_stripe;type:char(55);not null" json:"key_stripe"`
	KeyBraintree       	string             	`gorm:"column:key_braintree;type:char(55)" json:"key_braintree"`
	Type               	string             	`gorm:"column:type;not null" json:"type"`
	MaxUses            	int                	`gorm:"column:max_uses;type:int(10)" json:"max_uses"`
	MaxUsesUser        	int                	`gorm:"column:max_uses_user;type:int(10)" json:"max_uses_user"`
	Redeemable         	int8               	`gorm:"column:redeemable;type:tinyint(4)" json:"redeemable"`
	Approach           	string             	`gorm:"column:approach;type:enum('all','subscribed','not_subscribed');not null" json:"approach"`
	ExpireAt           	time.Time          	`gorm:"column:expire_at;type:datetime" json:"expire_at"`
	Period             	string             	`gorm:"column:period;type:char(255)" json:"period"`
}

// UserTrial [...]
type UserTrial struct{
	IDUser             	int                	`gorm:"primary_key;column:id_user;type:int(10) unsigned;not null" json:"-"`
	TokenCard          	string             	`gorm:"unique;column:token_card;type:varchar(64);not null" json:"token_card"`
	IDProduct          	int                	`gorm:"index:fk_id_product_user_trial;column:id_product;type:int(10) unsigned;not null" json:"id_product"`
	Product            	Product            	`gorm:"association_foreignkey:id_product;foreignkey:id_product" json:"prd_product_list"` // product label to assign price based on sales channel (country)
	DtStartTrial       	time.Time          	`gorm:"column:dt_start_trial;type:datetime" json:"dt_start_trial"`
	DtEndTrial         	time.Time          	`gorm:"column:dt_end_trial;type:datetime" json:"dt_end_trial"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"created_at"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"updated_at"`
	DeletedAt          	*time.Time         	`gorm:"column:deleted_at;type:datetime" json:"deleted_at"`
}


func (UserTrial)TableName() string {
	return "user_trial"
}


/*// SeasonTicket [...]
type SeasonTicket struct{
	IDSeasonTicket     	int                	`gorm:"primary_key;column:id_season_ticket;type:int(10);not null" json:"id_season_ticket"`
	IDUser             	uint               	`gorm:"column:id_user;type:int(10) unsigned;not null" json:"id_user"`
	IDProduct          	int                	`gorm:"index:fk_id_product_season_ticket;column:id_product;type:int(10) unsigned;not null" json:"id_product"`
	Type               	string             	`gorm:"column:type;type:enum('unique','subscription');not null" json:"type"`
	PrdProduct         	Product            	`gorm:"association_foreignkey:id_product;foreignkey:id_product" json:"prd_product_list"` // product label to assign price based on sales channel (country)
	FlTrial            	bool               	`gorm:"column:fl_trial;type:tinyint(1)" json:"fl_trial"`
	DtStartTrial       	*time.Time         	`gorm:"column:dt_start_trial;type:datetime" json:"dt_start_trial"`
	DtEndTrial         	*time.Time         	`gorm:"column:dt_end_trial;type:datetime" json:"dt_end_trial"`
	DtFirstPayment     	time.0Time         	`gorm:"column:dt_first_payment;type:datetime" json:"dt_first_payment"`
	DtLastPayment      	*time.Time         	`gorm:"column:dt_last_payment;type:datetime" json:"dt_last_payment"`
	DtNextPayment      	*time.Time         	`gorm:"column:dt_next_payment;type:datetime" json:"dt_next_payment"`
	PaymentGateway     	string             	`gorm:"column:payment_gateway;type:enum('stripe','paypal','braintree')" json:"payment_gateway"`
	BankToken          	string             	`gorm:"column:bank_token;type:varchar(64)" json:"bank_token"`
	DtEndSeasonTicket  	*time.Time         	`gorm:"column:dt_end_season_ticket;type:datetime" json:"dt_end_season_ticket"`
	DtCancellation     	*time.Time         	`gorm:"column:dt_cancellation;type:datetime" json:"dt_cancellation"`
	DtExpiration       	*time.Time         	`gorm:"column:dt_expiration;type:datetime" json:"dt_expiration"`
	NumEventBought     	int                	`gorm:"column:num_event_bought;type:int(10)" json:"num_event_bought"`
	NumEventConsumed   	int                	`gorm:"column:num_event_consumed;type:int(10)" json:"num_event_consumed"`
	PaymentFrequency   	string             	`gorm:"column:payment_frequency;type:enum('W','M','Y','1S');not null" json:"payment_frequency"`
	PaymentInterval    	uint8              	`gorm:"column:payment_interval;type:tinyint(3) unsigned" json:"payment_interval"`
	NumberOfPayments   	int                	`gorm:"column:number_of_payments;type:int(10);not null" json:"number_of_payments"`
	PaymentsDone       	int                	`gorm:"column:payments_done;type:int(10);not null" json:"payments_done"`
	PaymentPending     	int                	`gorm:"column:payment_pending;type:int(10);not null" json:"payment_pending"`
	LastPaymentOutcome 	string             	`gorm:"column:last_payment_outcome;type:enum('success','fail','retry');not null" json:"last_payment_outcome"`
	RetryCount         	bool               	`gorm:"column:retry_count;type:tinyint(1)" json:"retry_count"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"-"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"-"`
	DeletedAt          	*time.Time         	`gorm:"column:deleted_at;type:datetime" json:"-"`
}*/

// SeasonTicket A purchased product corresponds to a subscription (season ticket), the subscription summarizes all related information such as next payment, number of payments made, number of payments still to be made, expiration date, cancellation date, etc.
type SeasonTicket struct{
	IDSeasonTicket     	int                	`gorm:"primary_key;column:id_season_ticket;type:int(10);not null" json:"-"`
	IDUser             	uint               	`gorm:"column:id_user;type:int(10) unsigned;not null" json:"id_user"`
	IDProduct          	int                	`gorm:"index:fk_id_product_season_ticket;column:id_product;type:int(10) unsigned;not null" json:"id_product"` // product numeric identifier
	IDPgJob            	int                	`gorm:"column:id_pg_job;type:int(11);not null" json:"id_pg_job"`
	IDOrder            	int                	`gorm:"column:id_order;type:int(11);not null" json:"id_order"`
	Type               	string             	`gorm:"column:type;type:enum('subscription','ppv','carnet');not null" json:"type"`
	Trial              	bool               	`gorm:"column:trial;type:tinyint(1)" json:"trial"`
	StartTrial         	*time.Time         	`gorm:"column:start_trial;type:datetime" json:"start_trial"`
	EndTrial           	*time.Time         	`gorm:"column:end_trial;type:datetime" json:"end_trial"`
	//Cancellation     	*time.Time         	`gorm:"column:cancellation;type:datetime" json:"cancellation"`
	Expiration         	*time.Time         	`gorm:"column:expiration;type:datetime" json:"expiration"`
	NumEventBought     	int                	`gorm:"column:num_event_bought;type:int(10)" json:"num_event_bought"`
	NumEventConsumed   	int                	`gorm:"column:num_event_consumed;type:int(10)" json:"num_event_consumed"`
	Status             	string             	`gorm:"column:status;type:enum('active','trial','canceled','finished','overdue');not null" json:"status"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"created_at"` // Record creation date and time, if not specified in the insert list default value CURRENT_TIMESTAMP is used
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"updated_at"` // Record update date and time, if not specified in the insert list default value CURRENT_TIMESTAMP is used, if not specified in the update list default value CURRET_TIMESTAMP is used
	DeletedAt          	time.Time          	`gorm:"column:deleted_at;type:datetime" json:"deleted_at"`          // Record logical deletion of the record, if not specified in the insert list default value NULL is used, if not specified in the update list the field value in not modified, to logically delete the record update the field with CURRENT_TIMESTAMP value
}

func (SeasonTicket)TableName() string {
	return "season_ticket"
}

type LogViews struct{
	IDUser             	uint               	`gorm:"column:id_user;type:int(10) unsigned;not null" json:"id_user"`
	IDContent          	int                	`gorm:"column:id_content;" json:"id_content"`
	IDSeasonTicket     	int                	`gorm:"column:id_season_ticket;" json:"id_season_ticket"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"-"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"-"`
	DeletedAt          	*time.Time         	`gorm:"column:deleted_at;type:datetime" json:"-"`
}


func (LogViews)TableName() string {
	return "log_views"
}


type Subscription struct{
	Carnet             	int                	`json:"carnet"`
	Subscription       	[]SubscriptionItem 	`json:"subscription"`
}

type SubscriptionItem struct{
	Expire             	time.Time          	`gorm:"column_name:expire" json:"d"`
	Entity             	string             	`gorm:"column_name:entity" json:"e"`
	ID                 	int                	`gorm:"column_name:id" json:"i"`
}



type PgCustomer struct{
	IDUser             	int                	`gorm:"column:id_user;type:int(10);not null" json:"id_user"`
	IDCustomer         	int                	`gorm:"column:id_customer;type:int(11);not null" json:"id_customer"`
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:timestamp;not null" json:"created_at"`
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:timestamp;not null" json:"updated_at"`
	DeletedAt          	*time.Time         	`gorm:"column:deleted_at;type:timestamp" json:"deleted_at"`
}


func (PgCustomer)TableName() string {
	return "pg_customer"
}


type PrdComponent struct{
	IDComponent        	int                	`gorm:"primary_key;column:id_component;type:int(10);not null" json:"-"`                              // component id
	CodeComponent      	string             	`gorm:"unique;column:code_component;type:varchar(32);not null" json:"code_component"`                // component code (short name)
	NameComponent      	string             	`gorm:"unique;column:name_component;type:varchar(64);not null" json:"name_component"`                // component name (extended name/description); the product component name may be not unique
	DescComponent      	string             	`gorm:"column:desc_component;type:varchar(255);not null" json:"desc_component"`                      // component description, contains information about what the user with this role can do
	IDResource         	int                	`gorm:"index:idx_id_resource;column:id_resource;type:int(10) unsigned" json:"id_resource"`           // starting id in the table specified
	ResourceType       	string             	`gorm:"index:idx_resource_type;column:resource_type;type:varchar(30);not null" json:"resource_type"` // table specified
	CreatedAt          	time.Time          	`gorm:"column:created_at;type:datetime;not null" json:"created_at"`                                  // Record creation date and time, if not specified in the insert list default value CURRENT_TIMESTAMP is used
	UpdatedAt          	time.Time          	`gorm:"column:updated_at;type:datetime;not null" json:"updated_at"`                                  // Record update date and time, if not specified in the insert list default value CURRENT_TIMESTAMP is used, if not specified in the update list default value CURRET_TIMESTAMP is used
	DeletedAt          	time.Time          	`gorm:"column:deleted_at;type:datetime" json:"deleted_at"`                                           // Record logical deletion of the record, if not specified in the insert list default value NULL is used, if not specified in the update list the field value in not modified, to logically delete the record update the field with CURRENT_TIMESTAMP value
}

func (PrdComponent)TableName() string {
	return "prd_component"
}