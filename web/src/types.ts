export type Link = {
  id: number;
  slug: string;
  destination_url: string;
  title: string;
  is_custom: boolean;
  click_count: number;
  created_at: string;
  expires_at: string;
  status: Status;
  redirect_type: RedirectType;
};

export type Status = "active" | "disabled" | "expired";
export type RedirectType = "301" | "302";

export type LinkClicks = {
  id: number;
  link_id: number;
  referer: string;
  user_agent: string;
  clicked_at: string;
  browser: string;
  os: string;
  device: string;
  country: string;
  region: string;
  city: string;
  utm_params: UtmParams;
  qr_scan: boolean;
};

export type UtmParams = {};
