/**
 * odo dev
 * API interface for \'odo dev\'
 *
 * The version of the OpenAPI document: 0.1
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */


export interface DevstateImagePostRequest { 
    /**
     * Name of the image
     */
    name?: string;
    imageName?: string;
    args?: Array<string>;
    buildContext?: string;
    rootRequired?: boolean;
    uri?: string;
}

