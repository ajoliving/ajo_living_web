--
-- PostgreSQL database dump
--

\restrict zXy8LuIjam11UNvelJEaRqk41zpp1DgdFhA3uax5rXDP5cgq1R7Qkj0EJCX4emj

-- Dumped from database version 16.13
-- Dumped by pg_dump version 16.13

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: chat_participants; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chat_participants (
    id bigint NOT NULL,
    chat_id bigint NOT NULL,
    user_id bigint NOT NULL,
    role_in_chat character varying(32) NOT NULL,
    last_read_message_id bigint,
    unread_count bigint NOT NULL,
    joined_at timestamp with time zone
);


ALTER TABLE public.chat_participants OWNER TO postgres;

--
-- Name: chat_participants_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.chat_participants_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.chat_participants_id_seq OWNER TO postgres;

--
-- Name: chat_participants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.chat_participants_id_seq OWNED BY public.chat_participants.id;


--
-- Name: chats; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chats (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    biz_module character varying(32) NOT NULL,
    listing_id bigint NOT NULL,
    chat_type character varying(32) NOT NULL,
    created_by bigint NOT NULL,
    last_message_preview character varying(500),
    last_message_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.chats OWNER TO postgres;

--
-- Name: chats_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.chats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.chats_id_seq OWNER TO postgres;

--
-- Name: chats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.chats_id_seq OWNED BY public.chats.id;


--
-- Name: communities; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.communities (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    community_type character varying(32) NOT NULL,
    name_zh character varying(200) NOT NULL,
    name_en character varying(200),
    district_code character varying(32) NOT NULL,
    parent_community_id bigint,
    address_text character varying(500),
    created_at timestamp with time zone
);


ALTER TABLE public.communities OWNER TO postgres;

--
-- Name: communities_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.communities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.communities_id_seq OWNER TO postgres;

--
-- Name: communities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.communities_id_seq OWNED BY public.communities.id;


--
-- Name: contact_access_logs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.contact_access_logs (
    id bigint NOT NULL,
    listing_id bigint NOT NULL,
    request_user_id bigint NOT NULL,
    granted_channels jsonb NOT NULL,
    request_ip character varying(64),
    user_agent character varying(500),
    created_at timestamp with time zone
);


ALTER TABLE public.contact_access_logs OWNER TO postgres;

--
-- Name: contact_access_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.contact_access_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.contact_access_logs_id_seq OWNER TO postgres;

--
-- Name: contact_access_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.contact_access_logs_id_seq OWNED BY public.contact_access_logs.id;


--
-- Name: discover_placements; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.discover_placements (
    id bigint NOT NULL,
    scene character varying(64) NOT NULL,
    category_code character varying(64) DEFAULT ''::character varying NOT NULL,
    slot_index bigint NOT NULL,
    listing_id bigint NOT NULL,
    created_by bigint,
    updated_by bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.discover_placements OWNER TO postgres;

--
-- Name: discover_placements_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.discover_placements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.discover_placements_id_seq OWNER TO postgres;

--
-- Name: discover_placements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.discover_placements_id_seq OWNED BY public.discover_placements.id;


--
-- Name: home_content_placements; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.home_content_placements (
    id bigint NOT NULL,
    placement_type character varying(32) NOT NULL,
    module_code character varying(64) DEFAULT ''::character varying NOT NULL,
    slot_index bigint NOT NULL,
    media_asset_id bigint NOT NULL,
    title character varying(160),
    subtitle character varying(240),
    body text,
    created_by bigint,
    updated_by bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.home_content_placements OWNER TO postgres;

--
-- Name: home_content_placements_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.home_content_placements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.home_content_placements_id_seq OWNER TO postgres;

--
-- Name: home_content_placements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.home_content_placements_id_seq OWNED BY public.home_content_placements.id;


--
-- Name: listing_contacts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.listing_contacts (
    listing_id bigint NOT NULL,
    phone_encrypted text,
    phone_masked character varying(64),
    whats_app_encrypted text,
    whats_app_masked character varying(64),
    email_encrypted text,
    show_phone boolean NOT NULL,
    show_whats_app boolean NOT NULL,
    show_chat boolean NOT NULL,
    show_inquiry_form boolean NOT NULL,
    contact_mode character varying(32) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.listing_contacts OWNER TO postgres;

--
-- Name: listing_contacts_listing_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.listing_contacts_listing_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.listing_contacts_listing_id_seq OWNER TO postgres;

--
-- Name: listing_contacts_listing_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.listing_contacts_listing_id_seq OWNED BY public.listing_contacts.listing_id;


--
-- Name: listing_images; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.listing_images (
    id bigint NOT NULL,
    listing_id bigint NOT NULL,
    media_asset_id bigint NOT NULL,
    sort_order bigint NOT NULL,
    is_cover boolean NOT NULL,
    created_at timestamp with time zone
);


ALTER TABLE public.listing_images OWNER TO postgres;

--
-- Name: listing_images_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.listing_images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.listing_images_id_seq OWNER TO postgres;

--
-- Name: listing_images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.listing_images_id_seq OWNED BY public.listing_images.id;


--
-- Name: listings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.listings (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    module character varying(32) NOT NULL,
    owner_user_id bigint NOT NULL,
    title character varying(300) NOT NULL,
    summary character varying(500),
    description text,
    district_code character varying(32) NOT NULL,
    community_id bigint,
    publisher_identity_type character varying(32) NOT NULL,
    publication_status character varying(32) NOT NULL,
    moderation_status character varying(32) NOT NULL,
    business_status character varying(32) NOT NULL,
    published_at timestamp with time zone,
    sort_refreshed_at timestamp with time zone,
    expire_at timestamp with time zone,
    is_deleted boolean NOT NULL,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.listings OWNER TO postgres;

--
-- Name: listings_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.listings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.listings_id_seq OWNER TO postgres;

--
-- Name: listings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.listings_id_seq OWNED BY public.listings.id;


--
-- Name: media_assets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.media_assets (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    storage_provider character varying(32) NOT NULL,
    bucket_name character varying(120) NOT NULL,
    object_key character varying(500) NOT NULL,
    mime_type character varying(100) NOT NULL,
    width bigint,
    height bigint,
    file_size bigint,
    checksum_sha256 character varying(128),
    created_by bigint,
    created_at timestamp with time zone
);


ALTER TABLE public.media_assets OWNER TO postgres;

--
-- Name: media_assets_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.media_assets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.media_assets_id_seq OWNER TO postgres;

--
-- Name: media_assets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.media_assets_id_seq OWNED BY public.media_assets.id;


--
-- Name: messages; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.messages (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    chat_id bigint NOT NULL,
    sender_user_id bigint NOT NULL,
    message_type character varying(32) NOT NULL,
    content_text text NOT NULL,
    message_status character varying(32) NOT NULL,
    created_at timestamp with time zone,
    action_label character varying(80),
    action_url character varying(500)
);


ALTER TABLE public.messages OWNER TO postgres;

--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.messages_id_seq OWNER TO postgres;

--
-- Name: messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.messages_id_seq OWNED BY public.messages.id;


--
-- Name: moderation_actions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.moderation_actions (
    id bigint NOT NULL,
    target_type character varying(32) NOT NULL,
    target_id bigint NOT NULL,
    action_type character varying(32) NOT NULL,
    reason_code character varying(64),
    reason_text character varying(500),
    operator_user_id bigint NOT NULL,
    created_at timestamp with time zone
);


ALTER TABLE public.moderation_actions OWNER TO postgres;

--
-- Name: moderation_actions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.moderation_actions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.moderation_actions_id_seq OWNER TO postgres;

--
-- Name: moderation_actions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.moderation_actions_id_seq OWNED BY public.moderation_actions.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    user_id bigint NOT NULL,
    category character varying(64) NOT NULL,
    title character varying(200) NOT NULL,
    body character varying(500) NOT NULL,
    related_type character varying(64),
    related_public_id character varying(64),
    is_read boolean DEFAULT false NOT NULL,
    read_at timestamp with time zone,
    created_at timestamp with time zone
);


ALTER TABLE public.notifications OWNER TO postgres;

--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.notifications_id_seq OWNER TO postgres;

--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: order_logs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.order_logs (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    action_type character varying(64) NOT NULL,
    from_status character varying(32),
    to_status character varying(32),
    operator_user_id bigint NOT NULL,
    note character varying(500),
    created_at timestamp with time zone
);


ALTER TABLE public.order_logs OWNER TO postgres;

--
-- Name: order_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.order_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.order_logs_id_seq OWNER TO postgres;

--
-- Name: order_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.order_logs_id_seq OWNED BY public.order_logs.id;


--
-- Name: orders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.orders (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    biz_module character varying(32) NOT NULL,
    listing_id bigint NOT NULL,
    buyer_user_id bigint NOT NULL,
    seller_user_id bigint NOT NULL,
    order_status character varying(32) NOT NULL,
    buyer_note character varying(500),
    cancel_reason character varying(500),
    handover_method character varying(64),
    confirmed_at timestamp with time zone,
    completed_at timestamp with time zone,
    cancelled_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.orders OWNER TO postgres;

--
-- Name: orders_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.orders_id_seq OWNER TO postgres;

--
-- Name: orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.orders_id_seq OWNED BY public.orders.id;


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.permissions (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    code character varying(96) NOT NULL,
    scope character varying(32) NOT NULL,
    name character varying(160) NOT NULL,
    description character varying(500),
    is_system boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.permissions OWNER TO postgres;

--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.permissions_id_seq OWNER TO postgres;

--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.permissions_id_seq OWNED BY public.permissions.id;


--
-- Name: property_sale_listings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.property_sale_listings (
    listing_id bigint NOT NULL,
    property_type character varying(64),
    estate_name character varying(200),
    address_text character varying(500),
    asking_price_hkd numeric(14,2),
    usable_area_sqft bigint NOT NULL,
    gross_area_sqft bigint,
    bedroom_count bigint DEFAULT 0 NOT NULL,
    living_room_count bigint DEFAULT 0 NOT NULL,
    bathroom_count bigint DEFAULT 0 NOT NULL,
    floor_level character varying(64),
    direction character varying(64),
    building_age character varying(64),
    feature_tags jsonb,
    contact_method character varying(32) NOT NULL,
    publisher_role_label character varying(64)
);


ALTER TABLE public.property_sale_listings OWNER TO postgres;

--
-- Name: property_sale_listings_listing_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.property_sale_listings_listing_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.property_sale_listings_listing_id_seq OWNER TO postgres;

--
-- Name: property_sale_listings_listing_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.property_sale_listings_listing_id_seq OWNED BY public.property_sale_listings.listing_id;


--
-- Name: reward_ad_claims; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.reward_ad_claims (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    reward_ad_id bigint NOT NULL,
    user_id bigint NOT NULL,
    claim_date character varying(10) NOT NULL,
    status character varying(32) NOT NULL,
    watch_started_at timestamp with time zone,
    claimed_at timestamp with time zone,
    reward_points bigint NOT NULL,
    ip_address character varying(64),
    user_agent character varying(500),
    failure_reason character varying(300),
    wallet_transaction_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.reward_ad_claims OWNER TO postgres;

--
-- Name: reward_ad_claims_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.reward_ad_claims_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.reward_ad_claims_id_seq OWNER TO postgres;

--
-- Name: reward_ad_claims_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.reward_ad_claims_id_seq OWNED BY public.reward_ad_claims.id;


--
-- Name: reward_ads; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.reward_ads (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    title character varying(160) NOT NULL,
    summary character varying(500),
    cover_url character varying(800),
    media_url character varying(800),
    target_url character varying(800),
    reward_points bigint NOT NULL,
    watch_seconds bigint DEFAULT 30 NOT NULL,
    daily_user_limit bigint DEFAULT 1 NOT NULL,
    total_budget bigint DEFAULT 0 NOT NULL,
    total_granted bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    starts_at timestamp with time zone,
    ends_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    media_type character varying(32) DEFAULT 'image'::character varying NOT NULL
);


ALTER TABLE public.reward_ads OWNER TO postgres;

--
-- Name: reward_ads_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.reward_ads_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.reward_ads_id_seq OWNER TO postgres;

--
-- Name: reward_ads_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.reward_ads_id_seq OWNED BY public.reward_ads.id;


--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.role_permissions (
    role_id bigint NOT NULL,
    permission_id bigint NOT NULL,
    created_at timestamp with time zone
);


ALTER TABLE public.role_permissions OWNER TO postgres;

--
-- Name: roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.roles (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    code character varying(64) NOT NULL,
    scope character varying(32) NOT NULL,
    name character varying(120) NOT NULL,
    description character varying(500),
    is_system boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.roles OWNER TO postgres;

--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.roles_id_seq OWNER TO postgres;

--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- Name: secondhand_listings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.secondhand_listings (
    listing_id bigint NOT NULL,
    category_code character varying(64) NOT NULL,
    price_mode character varying(32) NOT NULL,
    price_hkd numeric(12,2),
    condition_level character varying(32) NOT NULL,
    dimension_text character varying(300),
    pickup_region_code character varying(32) NOT NULL,
    pickup_location_text character varying(300) NOT NULL,
    delivery_tags jsonb,
    visibility_scope character varying(32) NOT NULL,
    visible_community_id bigint,
    contact_method character varying(32) NOT NULL,
    is_free_giveaway boolean NOT NULL
);


ALTER TABLE public.secondhand_listings OWNER TO postgres;

--
-- Name: secondhand_listings_listing_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.secondhand_listings_listing_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.secondhand_listings_listing_id_seq OWNER TO postgres;

--
-- Name: secondhand_listings_listing_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.secondhand_listings_listing_id_seq OWNED BY public.secondhand_listings.listing_id;


--
-- Name: serviced_apartment_projects; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.serviced_apartment_projects (
    listing_id bigint NOT NULL,
    project_name character varying(200),
    lowest_monthly_rent_hkd numeric(12,2) NOT NULL,
    address_text character varying(500) NOT NULL,
    min_lease_months bigint DEFAULT 1 NOT NULL,
    facility_tags jsonb,
    service_tags jsonb,
    room_types jsonb,
    contact_method character varying(32) NOT NULL,
    publisher_role_label character varying(64)
);


ALTER TABLE public.serviced_apartment_projects OWNER TO postgres;

--
-- Name: serviced_apartment_projects_listing_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.serviced_apartment_projects_listing_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.serviced_apartment_projects_listing_id_seq OWNER TO postgres;

--
-- Name: serviced_apartment_projects_listing_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.serviced_apartment_projects_listing_id_seq OWNED BY public.serviced_apartment_projects.listing_id;


--
-- Name: user_credentials; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_credentials (
    user_id bigint NOT NULL,
    email character varying(255) NOT NULL,
    password_hash text,
    is_verified boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.user_credentials OWNER TO postgres;

--
-- Name: user_credentials_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_credentials_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_credentials_user_id_seq OWNER TO postgres;

--
-- Name: user_credentials_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_credentials_user_id_seq OWNED BY public.user_credentials.user_id;


--
-- Name: user_profiles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_profiles (
    user_id bigint NOT NULL,
    display_name character varying(120),
    publisher_identity_type character varying(32),
    primary_community_id bigint,
    district_code character varying(32),
    avatar_asset_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.user_profiles OWNER TO postgres;

--
-- Name: user_profiles_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_profiles_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_profiles_user_id_seq OWNER TO postgres;

--
-- Name: user_profiles_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_profiles_user_id_seq OWNED BY public.user_profiles.user_id;


--
-- Name: user_role_bindings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_role_bindings (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    role_id bigint NOT NULL,
    assigned_by bigint,
    assigned_at timestamp with time zone
);


ALTER TABLE public.user_role_bindings OWNER TO postgres;

--
-- Name: user_role_bindings_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_role_bindings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_role_bindings_id_seq OWNER TO postgres;

--
-- Name: user_role_bindings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_role_bindings_id_seq OWNED BY public.user_role_bindings.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    phone_country_code character varying(8) NOT NULL,
    phone_number character varying(32) NOT NULL,
    member_status character varying(32) NOT NULL,
    member_type character varying(32) DEFAULT 'user'::character varying NOT NULL,
    is_staff boolean DEFAULT false NOT NULL,
    is_verified_phone boolean NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: wallet_accounts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wallet_accounts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    balance bigint DEFAULT 0 NOT NULL,
    total_earned bigint DEFAULT 0 NOT NULL,
    total_spent bigint DEFAULT 0 NOT NULL,
    version bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.wallet_accounts OWNER TO postgres;

--
-- Name: wallet_accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.wallet_accounts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.wallet_accounts_id_seq OWNER TO postgres;

--
-- Name: wallet_accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.wallet_accounts_id_seq OWNED BY public.wallet_accounts.id;


--
-- Name: wallet_transactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wallet_transactions (
    id bigint NOT NULL,
    public_id character varying(26) NOT NULL,
    user_id bigint NOT NULL,
    direction character varying(16) NOT NULL,
    amount bigint NOT NULL,
    balance_before bigint NOT NULL,
    balance_after bigint NOT NULL,
    source_type character varying(64) NOT NULL,
    biz_module character varying(64) NOT NULL,
    action_type character varying(64) NOT NULL,
    listing_id bigint,
    reward_ad_id bigint,
    claim_id bigint,
    idempotency_key character varying(160) NOT NULL,
    operator_user_id bigint,
    note character varying(500),
    created_at timestamp with time zone
);


ALTER TABLE public.wallet_transactions OWNER TO postgres;

--
-- Name: wallet_transactions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.wallet_transactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.wallet_transactions_id_seq OWNER TO postgres;

--
-- Name: wallet_transactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.wallet_transactions_id_seq OWNED BY public.wallet_transactions.id;


--
-- Name: chat_participants id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chat_participants ALTER COLUMN id SET DEFAULT nextval('public.chat_participants_id_seq'::regclass);


--
-- Name: chats id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chats ALTER COLUMN id SET DEFAULT nextval('public.chats_id_seq'::regclass);


--
-- Name: communities id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.communities ALTER COLUMN id SET DEFAULT nextval('public.communities_id_seq'::regclass);


--
-- Name: contact_access_logs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.contact_access_logs ALTER COLUMN id SET DEFAULT nextval('public.contact_access_logs_id_seq'::regclass);


--
-- Name: discover_placements id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.discover_placements ALTER COLUMN id SET DEFAULT nextval('public.discover_placements_id_seq'::regclass);


--
-- Name: home_content_placements id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.home_content_placements ALTER COLUMN id SET DEFAULT nextval('public.home_content_placements_id_seq'::regclass);


--
-- Name: listing_contacts listing_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.listing_contacts ALTER COLUMN listing_id SET DEFAULT nextval('public.listing_contacts_listing_id_seq'::regclass);


--
-- Name: listing_images id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.listing_images ALTER COLUMN id SET DEFAULT nextval('public.listing_images_id_seq'::regclass);


--
-- Name: listings id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.listings ALTER COLUMN id SET DEFAULT nextval('public.listings_id_seq'::regclass);


--
-- Name: media_assets id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.media_assets ALTER COLUMN id SET DEFAULT nextval('public.media_assets_id_seq'::regclass);


--
-- Name: messages id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.messages ALTER COLUMN id SET DEFAULT nextval('public.messages_id_seq'::regclass);


--
-- Name: moderation_actions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.moderation_actions ALTER COLUMN id SET DEFAULT nextval('public.moderation_actions_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: order_logs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_logs ALTER COLUMN id SET DEFAULT nextval('public.order_logs_id_seq'::regclass);


--
-- Name: orders id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders ALTER COLUMN id SET DEFAULT nextval('public.orders_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: property_sale_listings listing_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.property_sale_listings ALTER COLUMN listing_id SET DEFAULT nextval('public.property_sale_listings_listing_id_seq'::regclass);


--
-- Name: reward_ad_claims id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reward_ad_claims ALTER COLUMN id SET DEFAULT nextval('public.reward_ad_claims_id_seq'::regclass);


--
-- Name: reward_ads id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reward_ads ALTER COLUMN id SET DEFAULT nextval('public.reward_ads_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- Name: secondhand_listings listing_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.secondhand_listings ALTER COLUMN listing_id SET DEFAULT nextval('public.secondhand_listings_listing_id_seq'::regclass);


--
-- Name: serviced_apartment_projects listing_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.serviced_apartment_projects ALTER COLUMN listing_id SET DEFAULT nextval('public.serviced_apartment_projects_listing_id_seq'::regclass);


--
-- Name: user_credentials user_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_credentials ALTER COLUMN user_id SET DEFAULT nextval('public.user_credentials_user_id_seq'::regclass);


--
-- Name: user_profiles user_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_profiles ALTER COLUMN user_id SET DEFAULT nextval('public.user_profiles_user_id_seq'::regclass);


--
-- Name: user_role_bindings id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_role_bindings ALTER COLUMN id SET DEFAULT nextval('public.user_role_bindings_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: wallet_accounts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wallet_accounts ALTER COLUMN id SET DEFAULT nextval('public.wallet_accounts_id_seq'::regclass);


--
-- Name: wallet_transactions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wallet_transactions ALTER COLUMN id SET DEFAULT nextval('public.wallet_transactions_id_seq'::regclass);


--
-- Data for Name: chat_participants; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.chat_participants (id, chat_id, user_id, role_in_chat, last_read_message_id, unread_count, joined_at) FROM stdin;
2	1	2	system	\N	0	2026-05-07 18:10:21.957409+08
1	1	1	recipient	5	0	2026-05-07 18:10:21.957409+08
6	3	2	system	\N	0	2026-05-10 01:07:32.817479+08
5	3	3	recipient	9	0	2026-05-10 01:07:32.817479+08
3	2	3	buyer	11	1	2026-05-10 01:07:32.553814+08
4	2	1	owner	12	0	2026-05-10 01:07:32.553815+08
\.


--
-- Data for Name: chats; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.chats (id, public_id, biz_module, listing_id, chat_type, created_by, last_message_preview, last_message_at, created_at, updated_at) FROM stdin;
1	01KR0YNFM5CMMAQHPT9DDK4R68	system	0	system_notice	1	看看是否能发通知呢📢	2026-05-07 18:10:21.966331+08	2026-05-07 18:10:21.957409+08	2026-05-07 18:10:21.966331+08
3	01KR6VASTH34Z242VW3ADX4BCZ	system	0	system_notice	3	最高100幣，可以當錢花	2026-05-10 01:07:32.817479+08	2026-05-10 01:07:32.817479+08	2026-05-10 01:07:32.817479+08
2	01KR6VASJ8A1HD60RB5C65Y15R	secondhand	1	direct_listing_chat	3	😀	2026-05-10 01:50:36.268212+08	2026-05-10 01:07:32.552299+08	2026-05-10 01:50:36.268212+08
\.


--
-- Data for Name: communities; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.communities (id, public_id, community_type, name_zh, name_en, district_code, parent_community_id, address_text, created_at) FROM stdin;
1	01KQC68PW67BKHZZHS8XYAXHSH	estate	康怡花園	Kornhill	hk_east	\N	Quarry Bay	2026-04-29 16:39:08.935166+08
2	01KQC68PW6QC5AW32RPNQJ85DM	building	太古城金星閣	Taikoo Shing Venus Mansion	hk_east	\N	Taikoo Shing	2026-04-29 16:39:08.935166+08
3	01KQC68PW6N57CRXEZG2Q75JKS	estate	麗港城	Laguna City	kwun_tong	\N	Lam Tin	2026-04-29 16:39:08.935166+08
\.


--
-- Data for Name: contact_access_logs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.contact_access_logs (id, listing_id, request_user_id, granted_channels, request_ip, user_agent, created_at) FROM stdin;
1	1	1	{"chat": true, "phone": false, "whatsapp": false, "inquiry_form": false}	127.0.0.1	Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36 Edg/147.0.0.0	2026-05-08 15:02:48.625314+08
2	1	1	{"chat": true, "phone": false, "whatsapp": false, "inquiry_form": false}	127.0.0.1	Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36 Edg/147.0.0.0	2026-05-08 15:02:50.074667+08
3	1	1	{"chat": true, "phone": false, "whatsapp": false, "inquiry_form": false}	127.0.0.1	Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36 Edg/147.0.0.0	2026-05-08 15:02:50.693311+08
\.


--
-- Data for Name: discover_placements; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.discover_placements (id, scene, category_code, slot_index, listing_id, created_by, updated_by, created_at, updated_at) FROM stdin;
1	discover_hero		1	1	1	1	2026-05-07 16:13:32.09996+08	2026-05-07 18:39:48.415722+08
5	discover_hero		3	1	1	1	2026-05-07 18:39:48.420672+08	2026-05-07 18:39:48.420672+08
2	discover_category_carousel	home_furniture	1	1	1	1	2026-05-07 16:14:26.669334+08	2026-05-07 18:39:48.423604+08
3	discover_category_carousel	home_furniture	2	1	1	1	2026-05-07 16:14:26.671482+08	2026-05-07 18:39:48.425253+08
4	discover_category_carousel	home_furniture	3	1	1	1	2026-05-07 16:14:26.673405+08	2026-05-07 18:39:48.426708+08
\.


--
-- Data for Name: home_content_placements; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.home_content_placements (id, placement_type, module_code, slot_index, media_asset_id, title, subtitle, body, created_by, updated_by, created_at, updated_at) FROM stdin;
1	login_hero		1	4	@jimmy teoh	Victoria Harbour, HK		1	1	2026-05-08 14:56:59.361278+08	2026-05-08 14:56:59.361278+08
2	login_hero		2	5	@kseniya kobi	Hong Kong Residence		1	1	2026-05-08 14:56:59.362529+08	2026-05-08 14:56:59.362529+08
3	carousel		1	6				1	1	2026-05-08 14:57:02.866732+08	2026-05-08 14:57:02.866732+08
4	carousel		2	7				1	1	2026-05-08 14:57:02.867165+08	2026-05-08 14:57:02.867165+08
5	carousel		3	8				1	1	2026-05-08 14:57:02.867483+08	2026-05-08 14:57:02.867483+08
6	module_card	secondhand	1	9	二手交易	已接入首頁、列表、篩選與聊天。	瀏覽屋苑二手帖子，支援篩選、聊天與成交跟進。	1	1	2026-05-08 14:57:03.967284+08	2026-05-08 14:57:03.967284+08
7	module_card	property_sale	2	10	樓盤放售	樓盤放售已可瀏覽與發布。	樓盤放售入口，支援列表、詳情、發布與聯絡方式解鎖。	1	1	2026-05-08 14:57:03.968473+08	2026-05-08 14:57:03.968473+08
8	module_card	serviced_apartment	3	11	服務住宅	服務式住宅已可瀏覽與發布。	短租與月租服務住宅入口，支援列表、詳情、發布與聯絡方式解鎖。	1	1	2026-05-08 14:57:03.968743+08	2026-05-08 14:57:03.968743+08
\.


--
-- Data for Name: listing_contacts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.listing_contacts (listing_id, phone_encrypted, phone_masked, whats_app_encrypted, whats_app_masked, email_encrypted, show_phone, show_whats_app, show_chat, show_inquiry_form, contact_mode, created_at, updated_at) FROM stdin;
1						f	f	t	f	chat_or_whatsapp	2026-04-29 17:08:57.792076+08	2026-04-29 17:08:57.791568+08
\.


--
-- Data for Name: listing_images; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.listing_images (id, listing_id, media_asset_id, sort_order, is_cover, created_at) FROM stdin;
1	1	2	1	t	2026-04-29 17:08:57.794024+08
\.


--
-- Data for Name: listings; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.listings (id, public_id, module, owner_user_id, title, summary, description, district_code, community_id, publisher_identity_type, publication_status, moderation_status, business_status, published_at, sort_refreshed_at, expire_at, is_deleted, deleted_at, created_at, updated_at) FROM stdin;
1	01KQC7Z9STR59G0SS7TW3N2DDB	secondhand	1	四層儲物櫃 九成新   高100 闊61 深48 cm	四層儲物櫃 九成新   高100 闊61 深48 cm	四層儲物櫃 九成新\n \n高100 闊61 深48 cm	eastern	\N	staff	active	approved	available	2026-04-29 17:08:57.812964+08	2026-04-29 17:08:57.812964+08	2026-05-13 17:08:57.812964+08	f	\N	2026-04-29 17:08:57.786107+08	2026-04-29 17:08:57.81321+08
\.


--
-- Data for Name: media_assets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.media_assets (id, public_id, storage_provider, bucket_name, object_key, mime_type, width, height, file_size, checksum_sha256, created_by, created_at) FROM stdin;
1	01KQC7Q6JE23D1SSV4HAMHGZSE	oss	ajo-living	ajo_living/01kqc7pw7zw8cwj065dvr7496v.jpg	image/jpeg	\N	\N	166220		1	2026-04-29 17:04:32.335447+08
2	01KQC7Z9S8Y2QAZ51853AT6J7T	oss	ajo-living	ajo_living/01kqc7z36h7akdb4tgny4yhk89.jpg	image/jpeg	\N	\N	166220		1	2026-04-29 17:08:57.769038+08
3	01KQC90BHCZG9PFRQZYCY2ZVKJ	oss	ajo-living	ajo_living/account/01kqc8zttwynyhxv7mbxn8h9nh.jpg	image/jpeg	\N	\N	4517223		1	2026-04-29 17:27:00.910334+08
4	01KR36027S8T91WAF127ES93B8	oss	ajo-living	ajo_living/login_bag/pexels-jimmy-teoh-294331-35774007.jpg	image/jpeg	2976	3968	1626925	90206aaba8816481f211caf97185de323e2d1846d2cff26ca97470d510b398bc	1	2026-05-08 14:56:57.595184+08
5	01KR3603YW686SB95JV9NBV2QH	oss	ajo-living	ajo_living/login_bag/pexels-kseniya-kobi-3624194-7820979.jpg	image/jpeg	3024	4032	1414656	72443f754d94112b1bd634cea6af118ee3dfb04beff9fba3f122681c9ba1e3ed	1	2026-05-08 14:56:59.357668+08
6	01KR3604A484BN6Y3S19GH4RKG	oss	ajo-living	ajo_living/eng/home-carousel/building.jpeg	image/jpeg	1280	720	77455	cb1f70124fcf10afed06877f733fd21f5ddb0c77b7d179421f5c26b0f01b24a2	1	2026-05-08 14:56:59.717621+08
7	01KR36060PMCBPWRR7DW4072XS	oss	ajo-living	ajo_living/eng/home-carousel/intercom.png	image/png	2730	1535	3383521	14a1fc4c3390370ce80aa584b3b59dc81034a7dc36604bf2bf16d3765bc44586	1	2026-05-08 14:57:01.463511+08
8	01KR3607CD91WDWNP1NYXZ31AS	oss	ajo-living	ajo_living/eng/home-carousel/rant.png	image/png	2730	1535	3116169	45598f8b9a32ea0ddcb3b336bdcbae9b0391062189f649908dea26f1e55b2bec	1	2026-05-08 14:57:02.863076+08
9	01KR3607R0147YR4TJ6YNQW1QT	oss	ajo-living	ajo_living/eng/home-carousel/secondhand.webp	image/webp	2816	1536	339960	d494a859b01f4482443b006b75ccb908abd4fe033689e6b51c45636f9ab7406d	1	2026-05-08 14:57:03.233415+08
10	01KR36083C27T89034G8CCN34H	oss	ajo-living	ajo_living/eng/home-carousel/property-sale.webp	image/webp	2816	1536	252792	9b6fb71f11fc2cbaaeb6e5fe2429c6ce71e5f56c875cfbf81eeaaa03b8781b56	1	2026-05-08 14:57:03.597682+08
11	01KR3608EVNPPBBJY8GW3WP3HG	oss	ajo-living	ajo_living/eng/home-carousel/serviced-apartment.webp	image/webp	2816	1536	109800	d99f465e4d3c7906f897cc25d35f1e49ae361e5852e01d8aa1f299498927b6ec	1	2026-05-08 14:57:03.964269+08
12	01KR3MFNGH2HC2W4CNK3R5HSPD	oss	ajo-living	ajo_living/advertisements/images/01kr3mfbdtaca18s466vqfjrx7.png	image/png	\N	\N	285577		1	2026-05-08 19:10:08.913515+08
13	01KR70AC52APHBKSF8RVSEYMK3	oss	ajo-living	ajo_living/advertisements/images/01kr70a7d88sbm71wef6pzshws.jpg	image/jpeg	\N	\N	4517223		1	2026-05-10 02:34:41.698969+08
\.


--
-- Data for Name: messages; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.messages (id, public_id, chat_id, sender_user_id, message_type, content_text, message_status, created_at, action_label, action_url) FROM stdin;
1	01KR0YNFMBY4ERCK1GJPDBX4QW	1	2	notice_card	分百萬紅包，領 IP 周邊\n水豚噜噜、讚萌 Loopy、奶龍等你投餵	sent	2026-05-07 18:05:21.957409+08	立即參與	/marketplace/discover
2	01KR0YNFMBFHRZ61EXSGKSRFET	1	2	notice_card	紅包到賬提醒\n拼手氣，瓜分 HK$35999 現金紅包	sent	2026-05-07 18:06:21.957409+08	去查看	/marketplace/discover
3	01KR0YNFMBXEYWRQT7V63SDPVH	1	2	notice_card	恭喜您獲得副業任務獎勵\n恭喜您本月賣出 1 筆副業寶貝，快去領取獎勵吧	sent	2026-05-07 18:07:21.957409+08	去領獎	/marketplace/my/orders
4	01KR0YNFMBEW5SXWYK87A07X9S	1	2	notice_card	閒魚幣登入獎勵到賬\n最高 100 幣，可以當錢花	sent	2026-05-07 18:08:21.957409+08	去兌換	/marketplace/discover
5	01KR0YNFMEABQC3ACYESZ9CS1E	1	2	notice_card	测试通知\n看看是否能发通知呢📢	sent	2026-05-07 18:10:21.966331+08	立即查看	http://localhost:5173/marketplace/settings
6	01KR6VASTMNYBGBXYW9YKX8TYK	3	2	notice_card	分百萬紅包，領 IP 周邊\n水豚噜噜、讚萌 Loopy、奶龍等你投餵	sent	2026-05-10 01:02:32.817479+08	立即參與	/marketplace/discover
7	01KR6VASTMC8DEX7MD202C20A9	3	2	notice_card	紅包到賬提醒\n拼手氣，瓜分 HK$35999 現金紅包	sent	2026-05-10 01:03:32.817479+08	去查看	/marketplace/discover
8	01KR6VASTMEPA2242KRGZ1JNCP	3	2	notice_card	恭喜您獲得副業任務獎勵\n恭喜您本月賣出 1 筆副業寶貝，快去領取獎勵吧	sent	2026-05-10 01:04:32.817479+08	去領獎	/marketplace/my/orders
9	01KR6VASTM96GG01V42V31V31B	3	2	notice_card	閒魚幣登入獎勵到賬\n最高 100 幣，可以當錢花	sent	2026-05-10 01:05:32.817479+08	去兌換	/marketplace/discover
10	01KR6VAY6HYDJ6KZFD8PPSBKDQ	2	3	text	你好	sent	2026-05-10 01:07:37.297606+08		
11	01KR6VBEYH0V2P01MHNYA8QV9Q	2	3	text	能收到信息么	sent	2026-05-10 01:07:54.449271+08		
12	01KR6XSMQ55DB48QBW3EEFADHA	2	1	text	😀	sent	2026-05-10 01:50:36.261881+08		
\.


--
-- Data for Name: moderation_actions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.moderation_actions (id, target_type, target_id, action_type, reason_code, reason_text, operator_user_id, created_at) FROM stdin;
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.notifications (id, public_id, user_id, category, title, body, related_type, related_public_id, is_read, read_at, created_at) FROM stdin;
1	01KR6VAY6RGGBKM8H8B25301XK	1	chat_message	New chat message	你好	chat	01KR6VASJ8A1HD60RB5C65Y15R	f	\N	2026-05-10 01:07:37.304361+08
2	01KR6VBEYKJ6V02V1CX00SASX7	1	chat_message	New chat message	能收到信息么	chat	01KR6VASJ8A1HD60RB5C65Y15R	f	\N	2026-05-10 01:07:54.451623+08
3	01KR6XSMQD766GHR8VAQM8R3FA	3	chat_message	New chat message	😀	chat	01KR6VASJ8A1HD60RB5C65Y15R	f	\N	2026-05-10 01:50:36.269867+08
\.


--
-- Data for Name: order_logs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.order_logs (id, order_id, action_type, from_status, to_status, operator_user_id, note, created_at) FROM stdin;
\.


--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.orders (id, public_id, biz_module, listing_id, buyer_user_id, seller_user_id, order_status, buyer_note, cancel_reason, handover_method, confirmed_at, completed_at, cancelled_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.permissions (id, public_id, code, scope, name, description, is_system, created_at, updated_at) FROM stdin;
5	01KQC68PT29QBB7PY7PAT05THX	order.create	member	Order Create	Create an order from a listing.	t	2026-04-29 16:39:08.866657+08	2026-05-10 02:15:14.894717+08
6	01KQC68PT2BCZ820ZNMP4KBC98	order.own.manage	member	Order Own Manage	Manage owned buyer or seller orders.	t	2026-04-29 16:39:08.866991+08	2026-05-10 02:15:14.895022+08
7	01KQC68PT3TC8Q87755DB8YFPV	notification.read	member	Notification Read	Read in-app notifications.	t	2026-04-29 16:39:08.867306+08	2026-05-10 02:15:14.895322+08
8	01KQC68PT37DQJVAKDEQ1EYQ8V	staff.console.access	staff	Staff Console Access	Access staff-only endpoints.	t	2026-04-29 16:39:08.867628+08	2026-05-10 02:15:14.895576+08
9	01KQC68PT38CQ9HV1WVEVFA420	staff.user.read	staff	Staff User Read	Read staff user list and summaries.	t	2026-04-29 16:39:08.867919+08	2026-05-10 02:15:14.895822+08
10	01KQC68PT4DM8EQTNZQA7BXXH3	staff.user.manage	staff	Staff User Manage	Manage member and staff access.	t	2026-04-29 16:39:08.868215+08	2026-05-10 02:15:14.896082+08
11	01KQC68PT4EVW9B0JZ726W45P5	staff.role.read	staff	Staff Role Read	Read available roles and permission matrices.	t	2026-04-29 16:39:08.868512+08	2026-05-10 02:15:14.896335+08
12	01KQC68PT4B63M226VXSP7RATF	staff.role.manage	staff	Staff Role Manage	Manage user role bindings.	t	2026-04-29 16:39:08.868797+08	2026-05-10 02:15:14.896635+08
13	01KQC68PT521DYG1S3SWEM6AR9	staff.review.manage	staff	Staff Review Manage	Handle moderation and review operations.	t	2026-04-29 16:39:08.86908+08	2026-05-10 02:15:14.896912+08
14	01KQC68PT5RDCTRTNTBW4PHGWR	staff.support.manage	staff	Staff Support Manage	Handle customer support and disputes.	t	2026-04-29 16:39:08.869361+08	2026-05-10 02:15:14.897202+08
1	01KQC68PT0J12G0Z4RSY8F74ZK	account.profile.read	member	Account Profile Read	Read current member profile.	t	2026-04-29 16:39:08.864688+08	2026-05-10 02:15:14.893124+08
2	01KQC68PT1T3W45YRQ7Z0JRA11	account.profile.write	member	Account Profile Write	Update current member profile.	t	2026-04-29 16:39:08.865544+08	2026-05-10 02:15:14.893721+08
3	01KQC68PT1ENBZTME4PGNEHSGA	listing.own.manage	member	Listing Own Manage	Manage owned secondhand listings.	t	2026-04-29 16:39:08.865939+08	2026-05-10 02:15:14.894064+08
4	01KQC68PT2N86HKXSNB9HCG7GW	chat.use	member	Chat Use	Use in-app listing chat.	t	2026-04-29 16:39:08.86629+08	2026-05-10 02:15:14.894358+08
\.


--
-- Data for Name: property_sale_listings; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.property_sale_listings (listing_id, property_type, estate_name, address_text, asking_price_hkd, usable_area_sqft, gross_area_sqft, bedroom_count, living_room_count, bathroom_count, floor_level, direction, building_age, feature_tags, contact_method, publisher_role_label) FROM stdin;
\.


--
-- Data for Name: reward_ad_claims; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.reward_ad_claims (id, public_id, reward_ad_id, user_id, claim_date, status, watch_started_at, claimed_at, reward_points, ip_address, user_agent, failure_reason, wallet_transaction_id, created_at, updated_at) FROM stdin;
1	01KR70AV3F0K10TM90DM5EBFBB	1	1	2026-05-10	claimed	2026-05-10 02:34:57.00682+08	2026-05-10 02:35:27.027064+08	50	127.0.0.1	Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36 Edg/147.0.0.0		1	2026-05-10 02:34:57.007728+08	2026-05-10 02:35:27.034843+08
\.


--
-- Data for Name: reward_ads; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.reward_ads (id, public_id, title, summary, cover_url, media_url, target_url, reward_points, watch_seconds, daily_user_limit, total_budget, total_granted, is_active, starts_at, ends_at, created_by, updated_by, created_at, updated_at, media_type) FROM stdin;
1	01KR70AC5KZ5826S48XHEKV5SQ	1	1		https://ajo-living.oss-cn-hongkong.aliyuncs.com/ajo_living/advertisements/images/01kr70a7d88sbm71wef6pzshws.jpg		50	30	1	0	50	t	2026-05-10 02:34:41.715061+08	2026-06-09 02:34:41.715061+08	1	1	2026-05-10 02:34:41.715061+08	2026-05-10 02:35:27.033636+08	image
\.


--
-- Data for Name: role_permissions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.role_permissions (role_id, permission_id, created_at) FROM stdin;
4	8	2026-04-29 16:39:08.878527+08
4	9	2026-04-29 16:39:08.878802+08
4	10	2026-04-29 16:39:08.879076+08
4	11	2026-04-29 16:39:08.879378+08
4	12	2026-04-29 16:39:08.879682+08
4	13	2026-04-29 16:39:08.88001+08
4	14	2026-04-29 16:39:08.880328+08
9	8	2026-05-07 13:40:00.792165+08
9	9	2026-05-07 13:40:00.792596+08
9	10	2026-05-07 13:40:00.792878+08
9	11	2026-05-07 13:40:00.79316+08
9	13	2026-05-07 13:40:00.793438+08
9	14	2026-05-07 13:40:00.793681+08
8	1	2026-05-07 13:40:00.794081+08
8	2	2026-05-07 13:40:00.79434+08
8	3	2026-05-07 13:40:00.794634+08
8	4	2026-05-07 13:40:00.794874+08
8	5	2026-05-07 13:40:00.795138+08
8	6	2026-05-07 13:40:00.795384+08
8	7	2026-05-07 13:40:00.795629+08
\.


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.roles (id, public_id, code, scope, name, description, is_system, created_at, updated_at) FROM stdin;
8	01KR0F6EMGQ9AEKPMH65Y7WKD8	member	member	Member	Baseline member access.	t	2026-05-07 13:40:00.784686+08	2026-05-10 02:15:14.897613+08
4	01KQC68PT77HN6BWTB9K7E5SEK	super_admin	staff	Super Admin	Full back-office access.	t	2026-04-29 16:39:08.8715+08	2026-05-10 02:15:14.898052+08
9	01KR0F6EMMCFA9TNVJM02YM7VC	staff	staff	Staff	Day-to-day back-office access.	t	2026-05-07 13:40:00.78848+08	2026-05-10 02:15:14.898361+08
\.


--
-- Data for Name: secondhand_listings; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.secondhand_listings (listing_id, category_code, price_mode, price_hkd, condition_level, dimension_text, pickup_region_code, pickup_location_text, delivery_tags, visibility_scope, visible_community_id, contact_method, is_free_giveaway) FROM stdin;
1	home_furniture	fixed	1198.00	used_good		eastern		[]	public	\N	chat_or_whatsapp	f
\.


--
-- Data for Name: serviced_apartment_projects; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.serviced_apartment_projects (listing_id, project_name, lowest_monthly_rent_hkd, address_text, min_lease_months, facility_tags, service_tags, room_types, contact_method, publisher_role_label) FROM stdin;
\.


--
-- Data for Name: user_credentials; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_credentials (user_id, email, password_hash, is_verified, created_at, updated_at) FROM stdin;
1	admin@admin.com	pbkdf2_sha256$120000$PmiEYLlN3kxzuF50s/QF0Q==$9KkByBJjS3qlE2huwN3DQpDkgayt0yhZSPtxNo5jiVg=	t	2026-04-29 16:39:08.93117+08	2026-05-10 02:15:14.940743+08
3	admin@admin.cn	pbkdf2_sha256$120000$RhyOaMdrHIzXs9VNypzRUw==$tX202UozDYYwex8bZsM8AzTN/KPPtaI/qQE4vwL2wyY=	t	2026-05-08 13:22:28.978811+08	2026-05-10 02:15:14.972505+08
\.


--
-- Data for Name: user_profiles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_profiles (user_id, display_name, publisher_identity_type, primary_community_id, district_code, avatar_asset_id, created_at, updated_at) FROM stdin;
1	Admin	staff	\N		3	2026-04-29 16:39:08.932195+08	2026-04-29 17:27:08.284446+08
3	Admin	staff	\N		\N	2026-05-08 13:22:28.984089+08	2026-05-08 13:22:28.984089+08
2	通知	system	\N		\N	2026-05-07 17:32:18.446448+08	2026-05-10 02:15:14.97451+08
\.


--
-- Data for Name: user_role_bindings; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_role_bindings (id, user_id, role_id, assigned_by, assigned_at) FROM stdin;
1	1	4	\N	2026-04-29 16:39:08.933547+08
2	3	4	\N	2026-05-08 13:22:28.984807+08
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, public_id, phone_country_code, phone_number, member_status, member_type, is_staff, is_verified_phone, created_at, updated_at) FROM stdin;
1	01KQC68PTK7QNTD78MREPYJ5R6	email	01KQC68PTKX7T9553671G732EW	active	user	t	f	2026-04-29 16:39:08.883976+08	2026-05-10 02:15:14.906412+08
3	01KR30K2EKCF164TXRG1VWBGR4	email	01KR30K2EK34Z7HMSTZY7NXFVE	active	user	t	f	2026-05-08 13:22:28.947193+08	2026-05-10 02:15:14.942529+08
2	01KR0WFSMDTB7SRH85CR0N1ETF	system	notification	active	user	f	f	2026-05-07 17:32:18.445479+08	2026-05-10 02:15:14.974076+08
\.


--
-- Data for Name: wallet_accounts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wallet_accounts (id, user_id, balance, total_earned, total_spent, version, created_at, updated_at) FROM stdin;
2	3	0	0	0	1	2026-05-10 01:07:29.377909+08	2026-05-10 01:07:29.377909+08
1	1	50	50	0	2	2026-05-08 17:04:42.463607+08	2026-05-10 02:35:27.033087+08
\.


--
-- Data for Name: wallet_transactions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wallet_transactions (id, public_id, user_id, direction, amount, balance_before, balance_after, source_type, biz_module, action_type, listing_id, reward_ad_id, claim_id, idempotency_key, operator_user_id, note, created_at) FROM stdin;
1	01KR70BRDQ3YZEYMHM5BDR6X8R	1	credit	50	0	50	reward_ad	wallet	ad_reward	\N	1	1	reward_ad:1:1:01KR70AV3F0K10TM90DM5EBFBB	\N	rewarded ad completed	2026-05-10 02:35:27.031392+08
\.


--
-- Name: chat_participants_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.chat_participants_id_seq', 6, true);


--
-- Name: chats_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.chats_id_seq', 3, true);


--
-- Name: communities_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.communities_id_seq', 3, true);


--
-- Name: contact_access_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.contact_access_logs_id_seq', 3, true);


--
-- Name: discover_placements_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.discover_placements_id_seq', 5, true);


--
-- Name: home_content_placements_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.home_content_placements_id_seq', 8, true);


--
-- Name: listing_contacts_listing_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.listing_contacts_listing_id_seq', 1, false);


--
-- Name: listing_images_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.listing_images_id_seq', 1, true);


--
-- Name: listings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.listings_id_seq', 1, true);


--
-- Name: media_assets_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.media_assets_id_seq', 13, true);


--
-- Name: messages_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.messages_id_seq', 12, true);


--
-- Name: moderation_actions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.moderation_actions_id_seq', 1, false);


--
-- Name: notifications_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notifications_id_seq', 3, true);


--
-- Name: order_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.order_logs_id_seq', 1, false);


--
-- Name: orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.orders_id_seq', 1, false);


--
-- Name: permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.permissions_id_seq', 14, true);


--
-- Name: property_sale_listings_listing_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.property_sale_listings_listing_id_seq', 1, false);


--
-- Name: reward_ad_claims_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.reward_ad_claims_id_seq', 1, true);


--
-- Name: reward_ads_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.reward_ads_id_seq', 1, true);


--
-- Name: roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.roles_id_seq', 9, true);


--
-- Name: secondhand_listings_listing_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.secondhand_listings_listing_id_seq', 1, false);


--
-- Name: serviced_apartment_projects_listing_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.serviced_apartment_projects_listing_id_seq', 1, false);


--
-- Name: user_credentials_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_credentials_user_id_seq', 1, false);


--
-- Name: user_profiles_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_profiles_user_id_seq', 1, false);


--
-- Name: user_role_bindings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_role_bindings_id_seq', 2, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 3, true);


--
-- Name: wallet_accounts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.wallet_accounts_id_seq', 2, true);


--
-- Name: wallet_transactions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.wallet_transactions_id_seq', 1, true);


--
-- Name: chat_participants chat_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chat_participants
    ADD CONSTRAINT chat_participants_pkey PRIMARY KEY (id);


--
-- Name: chats chats_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_pkey PRIMARY KEY (id);


--
-- Name: communities communities_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.communities
    ADD CONSTRAINT communities_pkey PRIMARY KEY (id);


--
-- Name: contact_access_logs contact_access_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.contact_access_logs
    ADD CONSTRAINT contact_access_logs_pkey PRIMARY KEY (id);


--
-- Name: discover_placements discover_placements_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.discover_placements
    ADD CONSTRAINT discover_placements_pkey PRIMARY KEY (id);


--
-- Name: home_content_placements home_content_placements_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.home_content_placements
    ADD CONSTRAINT home_content_placements_pkey PRIMARY KEY (id);


--
-- Name: listing_contacts listing_contacts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.listing_contacts
    ADD CONSTRAINT listing_contacts_pkey PRIMARY KEY (listing_id);


--
-- Name: listing_images listing_images_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.listing_images
    ADD CONSTRAINT listing_images_pkey PRIMARY KEY (id);


--
-- Name: listings listings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.listings
    ADD CONSTRAINT listings_pkey PRIMARY KEY (id);


--
-- Name: media_assets media_assets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.media_assets
    ADD CONSTRAINT media_assets_pkey PRIMARY KEY (id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: moderation_actions moderation_actions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.moderation_actions
    ADD CONSTRAINT moderation_actions_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: order_logs order_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_logs
    ADD CONSTRAINT order_logs_pkey PRIMARY KEY (id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: property_sale_listings property_sale_listings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.property_sale_listings
    ADD CONSTRAINT property_sale_listings_pkey PRIMARY KEY (listing_id);


--
-- Name: reward_ad_claims reward_ad_claims_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reward_ad_claims
    ADD CONSTRAINT reward_ad_claims_pkey PRIMARY KEY (id);


--
-- Name: reward_ads reward_ads_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reward_ads
    ADD CONSTRAINT reward_ads_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: secondhand_listings secondhand_listings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.secondhand_listings
    ADD CONSTRAINT secondhand_listings_pkey PRIMARY KEY (listing_id);


--
-- Name: serviced_apartment_projects serviced_apartment_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.serviced_apartment_projects
    ADD CONSTRAINT serviced_apartment_projects_pkey PRIMARY KEY (listing_id);


--
-- Name: user_credentials user_credentials_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_credentials
    ADD CONSTRAINT user_credentials_pkey PRIMARY KEY (user_id);


--
-- Name: user_profiles user_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (user_id);


--
-- Name: user_role_bindings user_role_bindings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_role_bindings
    ADD CONSTRAINT user_role_bindings_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: wallet_accounts wallet_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wallet_accounts
    ADD CONSTRAINT wallet_accounts_pkey PRIMARY KEY (id);


--
-- Name: wallet_transactions wallet_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_pkey PRIMARY KEY (id);


--
-- Name: idx_chat_participants_chat_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_chat_participants_chat_id ON public.chat_participants USING btree (chat_id);


--
-- Name: idx_chat_participants_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_chat_participants_user_id ON public.chat_participants USING btree (user_id);


--
-- Name: idx_chats_created_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_chats_created_by ON public.chats USING btree (created_by);


--
-- Name: idx_chats_listing_updated; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_chats_listing_updated ON public.chats USING btree (listing_id, updated_at);


--
-- Name: idx_chats_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_chats_public_id ON public.chats USING btree (public_id);


--
-- Name: idx_communities_district_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_communities_district_code ON public.communities USING btree (district_code);


--
-- Name: idx_communities_name_zh; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_communities_name_zh ON public.communities USING btree (name_zh);


--
-- Name: idx_communities_parent_community_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_communities_parent_community_id ON public.communities USING btree (parent_community_id);


--
-- Name: idx_communities_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_communities_public_id ON public.communities USING btree (public_id);


--
-- Name: idx_contact_access_logs_listing_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_contact_access_logs_listing_id ON public.contact_access_logs USING btree (listing_id);


--
-- Name: idx_contact_access_logs_request_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_contact_access_logs_request_user_id ON public.contact_access_logs USING btree (request_user_id);


--
-- Name: idx_discover_placements_category_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_discover_placements_category_code ON public.discover_placements USING btree (category_code);


--
-- Name: idx_discover_placements_created_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_discover_placements_created_by ON public.discover_placements USING btree (created_by);


--
-- Name: idx_discover_placements_listing_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_discover_placements_listing_id ON public.discover_placements USING btree (listing_id);


--
-- Name: idx_discover_placements_scene; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_discover_placements_scene ON public.discover_placements USING btree (scene);


--
-- Name: idx_discover_placements_updated_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_discover_placements_updated_by ON public.discover_placements USING btree (updated_by);


--
-- Name: idx_home_content_placements_created_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_home_content_placements_created_by ON public.home_content_placements USING btree (created_by);


--
-- Name: idx_home_content_placements_media_asset_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_home_content_placements_media_asset_id ON public.home_content_placements USING btree (media_asset_id);


--
-- Name: idx_home_content_placements_module_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_home_content_placements_module_code ON public.home_content_placements USING btree (module_code);


--
-- Name: idx_home_content_placements_placement_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_home_content_placements_placement_type ON public.home_content_placements USING btree (placement_type);


--
-- Name: idx_home_content_placements_updated_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_home_content_placements_updated_by ON public.home_content_placements USING btree (updated_by);


--
-- Name: idx_listing_images_listing_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_listing_images_listing_id ON public.listing_images USING btree (listing_id);


--
-- Name: idx_listing_images_media_asset_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_listing_images_media_asset_id ON public.listing_images USING btree (media_asset_id);


--
-- Name: idx_listings_community_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_listings_community_id ON public.listings USING btree (community_id);


--
-- Name: idx_listings_expire_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_listings_expire_at ON public.listings USING btree (publication_status, expire_at);


--
-- Name: idx_listings_is_deleted; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_listings_is_deleted ON public.listings USING btree (is_deleted);


--
-- Name: idx_listings_module_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_listings_module_status ON public.listings USING btree (module, publication_status, moderation_status, sort_refreshed_at);


--
-- Name: idx_listings_owner_module_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_listings_owner_module_status ON public.listings USING btree (owner_user_id, publication_status);


--
-- Name: idx_listings_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_listings_public_id ON public.listings USING btree (public_id);


--
-- Name: idx_media_assets_created_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_media_assets_created_by ON public.media_assets USING btree (created_by);


--
-- Name: idx_media_assets_object_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_media_assets_object_key ON public.media_assets USING btree (object_key);


--
-- Name: idx_media_assets_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_media_assets_public_id ON public.media_assets USING btree (public_id);


--
-- Name: idx_messages_chat_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_messages_chat_time ON public.messages USING btree (chat_id, created_at);


--
-- Name: idx_messages_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_messages_public_id ON public.messages USING btree (public_id);


--
-- Name: idx_messages_sender_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_messages_sender_user_id ON public.messages USING btree (sender_user_id);


--
-- Name: idx_moderation_actions_operator_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_moderation_actions_operator_user_id ON public.moderation_actions USING btree (operator_user_id);


--
-- Name: idx_moderation_actions_target_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_moderation_actions_target_id ON public.moderation_actions USING btree (target_id);


--
-- Name: idx_notifications_category; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_notifications_category ON public.notifications USING btree (category);


--
-- Name: idx_notifications_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_notifications_public_id ON public.notifications USING btree (public_id);


--
-- Name: idx_notifications_user_read_created; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_notifications_user_read_created ON public.notifications USING btree (user_id, is_read, created_at);


--
-- Name: idx_order_logs_operator_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_order_logs_operator_user_id ON public.order_logs USING btree (operator_user_id);


--
-- Name: idx_order_logs_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_order_logs_order_id ON public.order_logs USING btree (order_id);


--
-- Name: idx_orders_biz_module; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_biz_module ON public.orders USING btree (biz_module);


--
-- Name: idx_orders_buyer_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_buyer_user_id ON public.orders USING btree (buyer_user_id);


--
-- Name: idx_orders_listing_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_listing_id ON public.orders USING btree (listing_id);


--
-- Name: idx_orders_order_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_order_status ON public.orders USING btree (order_status);


--
-- Name: idx_orders_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_orders_public_id ON public.orders USING btree (public_id);


--
-- Name: idx_orders_seller_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_seller_user_id ON public.orders USING btree (seller_user_id);


--
-- Name: idx_permissions_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_permissions_code ON public.permissions USING btree (code);


--
-- Name: idx_permissions_is_system; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_permissions_is_system ON public.permissions USING btree (is_system);


--
-- Name: idx_permissions_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_permissions_public_id ON public.permissions USING btree (public_id);


--
-- Name: idx_permissions_scope; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_permissions_scope ON public.permissions USING btree (scope);


--
-- Name: idx_reward_ad_claims_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_reward_ad_claims_public_id ON public.reward_ad_claims USING btree (public_id);


--
-- Name: idx_reward_ad_claims_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reward_ad_claims_status ON public.reward_ad_claims USING btree (status);


--
-- Name: idx_reward_ad_claims_wallet_transaction_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reward_ad_claims_wallet_transaction_id ON public.reward_ad_claims USING btree (wallet_transaction_id);


--
-- Name: idx_reward_ads_created_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reward_ads_created_by ON public.reward_ads USING btree (created_by);


--
-- Name: idx_reward_ads_ends_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reward_ads_ends_at ON public.reward_ads USING btree (ends_at);


--
-- Name: idx_reward_ads_is_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reward_ads_is_active ON public.reward_ads USING btree (is_active);


--
-- Name: idx_reward_ads_media_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reward_ads_media_type ON public.reward_ads USING btree (media_type);


--
-- Name: idx_reward_ads_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_reward_ads_public_id ON public.reward_ads USING btree (public_id);


--
-- Name: idx_reward_ads_starts_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reward_ads_starts_at ON public.reward_ads USING btree (starts_at);


--
-- Name: idx_reward_ads_updated_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reward_ads_updated_by ON public.reward_ads USING btree (updated_by);


--
-- Name: idx_reward_claim_daily; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reward_claim_daily ON public.reward_ad_claims USING btree (claim_date);


--
-- Name: idx_reward_claim_user_ad; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reward_claim_user_ad ON public.reward_ad_claims USING btree (user_id, reward_ad_id);


--
-- Name: idx_roles_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_roles_code ON public.roles USING btree (code);


--
-- Name: idx_roles_is_system; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_roles_is_system ON public.roles USING btree (is_system);


--
-- Name: idx_roles_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_roles_public_id ON public.roles USING btree (public_id);


--
-- Name: idx_roles_scope; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_roles_scope ON public.roles USING btree (scope);


--
-- Name: idx_secondhand_visibility_community; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_secondhand_visibility_community ON public.secondhand_listings USING btree (visibility_scope, visible_community_id);


--
-- Name: idx_user_credentials_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_user_credentials_email ON public.user_credentials USING btree (email);


--
-- Name: idx_user_profiles_primary_community_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_profiles_primary_community_id ON public.user_profiles USING btree (primary_community_id);


--
-- Name: idx_user_role_bindings_assigned_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_role_bindings_assigned_by ON public.user_role_bindings USING btree (assigned_by);


--
-- Name: idx_user_role_bindings_role_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_role_bindings_role_id ON public.user_role_bindings USING btree (role_id);


--
-- Name: idx_user_role_bindings_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_role_bindings_user_id ON public.user_role_bindings USING btree (user_id);


--
-- Name: idx_users_is_staff; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_is_staff ON public.users USING btree (is_staff);


--
-- Name: idx_users_member_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_member_status ON public.users USING btree (member_status);


--
-- Name: idx_users_member_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_member_type ON public.users USING btree (member_type);


--
-- Name: idx_users_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_users_public_id ON public.users USING btree (public_id);


--
-- Name: idx_wallet_accounts_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_wallet_accounts_user_id ON public.wallet_accounts USING btree (user_id);


--
-- Name: idx_wallet_transactions_action_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wallet_transactions_action_type ON public.wallet_transactions USING btree (action_type);


--
-- Name: idx_wallet_transactions_biz_module; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wallet_transactions_biz_module ON public.wallet_transactions USING btree (biz_module);


--
-- Name: idx_wallet_transactions_claim_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wallet_transactions_claim_id ON public.wallet_transactions USING btree (claim_id);


--
-- Name: idx_wallet_transactions_direction; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wallet_transactions_direction ON public.wallet_transactions USING btree (direction);


--
-- Name: idx_wallet_transactions_idempotency_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_wallet_transactions_idempotency_key ON public.wallet_transactions USING btree (idempotency_key);


--
-- Name: idx_wallet_transactions_listing_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wallet_transactions_listing_id ON public.wallet_transactions USING btree (listing_id);


--
-- Name: idx_wallet_transactions_operator_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wallet_transactions_operator_user_id ON public.wallet_transactions USING btree (operator_user_id);


--
-- Name: idx_wallet_transactions_public_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_wallet_transactions_public_id ON public.wallet_transactions USING btree (public_id);


--
-- Name: idx_wallet_transactions_reward_ad_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wallet_transactions_reward_ad_id ON public.wallet_transactions USING btree (reward_ad_id);


--
-- Name: idx_wallet_transactions_source_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wallet_transactions_source_type ON public.wallet_transactions USING btree (source_type);


--
-- Name: idx_wallet_transactions_user_created; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wallet_transactions_user_created ON public.wallet_transactions USING btree (user_id, created_at);


--
-- Name: uk_chat_participants_chat_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uk_chat_participants_chat_user ON public.chat_participants USING btree (chat_id, user_id);


--
-- Name: uk_discover_placements_position; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uk_discover_placements_position ON public.discover_placements USING btree (scene, category_code, slot_index);


--
-- Name: uk_home_content_position; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uk_home_content_position ON public.home_content_placements USING btree (placement_type, module_code, slot_index);


--
-- Name: uk_user_phone; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uk_user_phone ON public.users USING btree (phone_country_code, phone_number);


--
-- Name: uk_user_role_bindings_user_role; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uk_user_role_bindings_user_role ON public.user_role_bindings USING btree (user_id, role_id);


--
-- Name: communities fk_communities_parent_community; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.communities
    ADD CONSTRAINT fk_communities_parent_community FOREIGN KEY (parent_community_id) REFERENCES public.communities(id);


--
-- Name: reward_ad_claims fk_reward_ad_claims_reward_ad; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reward_ad_claims
    ADD CONSTRAINT fk_reward_ad_claims_reward_ad FOREIGN KEY (reward_ad_id) REFERENCES public.reward_ads(id);


--
-- Name: user_credentials fk_user_credentials_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_credentials
    ADD CONSTRAINT fk_user_credentials_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: user_profiles fk_user_profiles_primary_community; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT fk_user_profiles_primary_community FOREIGN KEY (primary_community_id) REFERENCES public.communities(id);


--
-- PostgreSQL database dump complete
--

\unrestrict zXy8LuIjam11UNvelJEaRqk41zpp1DgdFhA3uax5rXDP5cgq1R7Qkj0EJCX4emj

