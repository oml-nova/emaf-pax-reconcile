function getCardNetworkType(abbreviation) {
    switch (abbreviation) {
        case 'AMEX':
            return 'American Express'
        case 'DISC':
            return 'Discover Network';
        case 'DSCV':
            return 'Discover Network';
        case 'FCOR':
            return 'FleetCor/Fuelman Auth';
        case 'FONE':
            return 'FleetOne Auth';
        case 'MCRD':
            return 'MasterCard';
        case 'PRVT':
            return 'Private Label';
        case 'PLCB':
            return 'Private Label – Cobrand';
        case 'VISA':
            return 'Visa';
        case 'VYGR':
            return 'Voyager';
        case 'WEX':
            return 'Wright Express';
        case 'STAR':
            return 'Star West';
        case 'MNY1':
            return 'Cash Station';
        case 'RCPA':
            return 'Star East Reciprocal';
        case 'AVAL':
            return 'Star East';
        case 'MAC2':
            return 'Star Northeast';
        case 'MAC3':
            return 'Star Northeast';
        case 'EXPL':
            return 'Explore/Star';
        case 'BKMT':
            return 'Star Midwest';
        case 'MACG':
            return 'MAC/Gr Machine';
        case 'MMAC':
            return 'MAC';
        case 'HNMN':
            return 'Star East';
        case 'CIXT':
            return 'Cash St Issuer';
        case 'MC3D':
            return 'MAC Conversion';
        case 'EXSS':
            return 'Explore Subswitch';
        case 'SESS':
            return 'Star East Subswitch';
        case 'SWSS':
            return 'Star West Subswitch';
        case 'EXP3':
            return 'Star West';
        case 'EXP4':
            return 'Star West';
        case 'STGY':
            return 'Star West Gateway';
        case 'SN1G':
            return 'Star NE POS 1 Groc';
        case 'SN1N':
            return 'Star NE POS 1 Non-Groc';
        case 'SN2G':
            return 'Star NE POS 2 Groc';
        case 'SN2N':
            return 'Star NE POS 2 Non-Groc';
        case 'SN3G':
            return 'Star NE POS 3 Groc';
        case 'SN3N':
            return 'Star NE POS 3 Non-Groc';
        case 'SS1G':
            return 'Star W POS 1 Groc';
        case 'SS1N':
            return 'Star SE POS 1 Non-Groc';
        case 'SS2G':
            return 'Star SE POS 2 Groc';
        case 'SS2N':
            return 'Star SE POS 2 Non-Groc';
        case 'SS3G':
            return 'Star SE POS 3 Groc';
        case 'SS3N':
            return 'Star SE POS 3 Non-Groc';
        case 'SW1G':
            return 'Star W POS 1 Groc';
        case 'SW1N':
            return 'Star W POS 1 Non-Groc';
        case 'SW2G':
            return 'Star W POS 2 Groc';
        case 'SW2N':
            return 'Star W POS 2 Non-Groc';
        case 'SW3G':
            return 'Star W POS 3 Groc';
        case 'SW3N':
            return 'Star W POS 3 Non-Groc';
        case 'MAC1':
            return 'MAC';
        case 'SEST':
            return 'Star East';
        case 'EXP6':
            return 'Star West';
        case 'EXP5':
            return 'Star West';
        case 'SN1P':
            return 'Star NE POS 1 Petro';
        case 'SN2P':
            return 'Star NE POS 2 Petro';
        case 'SN3P':
            return 'Star NE POS 3 Petro';
        case 'SS1P':
            return 'Star SE POS 1 Petro';
        case 'SS2P':
            return 'Star SE POS 2 Petro';
        case 'SS3P':
            return 'Star SE POS 3 Petro';
        case 'SW1P':
            return 'Star W POS 1 Petro';
        case 'SW2P':
            return 'Star W POS 2 Petro';
        case 'SW3P':
            return 'Star W POS 3 Petro';
        case 'SN1B':
            return 'Star NE POS 1 AP BL Pay';
        case 'SN2B':
            return 'Star NE POS 2 AP BL Pay';
        case 'SN3B':
            return 'Star NE POS 3 AP BL Pay';
        case 'SS1B':
            return 'Star SE POS 1 AP BL Pay';
        case 'SS2B':
            return 'Star SE POS 2 AP BL Pay';
        case 'SS3B':
            return 'Star SE POS 3 AP BL Pay';
        case 'SW1B':
            return 'Star W POS 1 AP BL Pay';
        case 'SW2B':
            return 'Star W POS 2 AP BL Pay';
        case 'SW3B':
            return 'Star W POS 3 AP BL Pay';
        case 'SNSB':
            return 'Star NE BP Stnd';
        case 'SSSB':
            return 'Star SE BP Stnd';
        case 'SWSB':
            return 'Star W BP Stnd';
        case 'SNEB':
            return 'Star NE BP Emerging';
        case 'SSEB':
            return 'Star SE BP Emerging';
        case 'SWEB':
            return 'Star W BP Emerging';
        case 'SNST':
            return 'Star NE POS Small Ticket';
        case 'SSST':
            return 'Star SE POS Small Ticket';
        case 'SWST':
            return 'Star W POS Small Ticket';
        case 'SNMR':
            return 'Star NE POS Medical';
        case 'SSMR':
            return 'Star SE POS Medical';
        case 'SWMR':
            return 'Star W POS Medical';
        case 'INLK':
            return 'Interlink';
        case 'INT2':
            return 'Interlink';
        case 'INS1':
            return 'Interlink Stnd SMF1';
        case 'INS2':
            return 'Interlink Stnd SMF2';
        case 'INS3':
            return 'Interlink Stnd SMF3';
        case 'ING1':
            return 'Interlink Groc SMF1';
        case 'ING2':
            return 'Interlink Groc SMF2';
        case 'ING3':
            return 'Interlink Groc SMF3';
        case 'PUL2':
            return 'Pulse';
        case 'TYME':
            return 'TYME';
        case 'TYM2':
            return 'TYME 2';
        case 'MS1':
            return 'Pulse -Money Station';
        case 'PULP':
            return 'Pulse';
        case 'MSZA':
            return 'Pulse Zone A';
        case 'MSZB':
            return 'Pulse Zone B';
        case 'MONY':
            return 'Money Station USB';
        case 'MSRN':
            return 'Pulse Reciprocal NYCE';
        case 'MSRS':
            return 'Pulse Reciprocal Star';
        case 'MSRT':
            return 'Pulse Reciprocal TYME';
        case 'MS9A':
            return 'Pulse Zone A Intsw';
        case 'MS9B':
            return 'Pulse Zone B Intsw';
        case 'PUZA':
            return 'Pulse Gateway Zone A';
        case 'PUZB':
            return 'Pulse Gateway Zone B';
        case 'PURN':
            return 'Pulse Gtwy Rcip NYCE';
        case 'PURS':
            return 'Pulse Gtwy Rcip Star';
        case 'PURT':
            return 'Pulse Gtwy Rcip TYME';
        case 'PU9A':
            return 'Pulse Gtwy A Intsw';
        case 'PU9B':
            return 'Pulse Gtwy B Intsw';
        case 'PNSR':
            return 'Pulse Limited Recip';
        case 'MNSR':
            return 'Pulse-MSI Ltd Recip';
        case 'PSI2':
            return 'Plus Network';
        case 'NYC1':
            return 'NYCE East';
        case 'MGCL':
            return 'NYCE Midwest';
        case 'JENI':
            return 'Jeanie';
        case 'JPIX':
            return 'Exempt Jeanie POS';
        case 'JEPR':
            return 'Jeanie Presto';
        case 'JENA':
            return 'Jeanie Audio Network';
        case 'MJI1':
            return "MPS Jeanie Int'l";
        case 'MJN1':
            return 'Jeanie RIC';
        case 'JEST':
            return 'Jeanie Northeast';
        case 'JNIT':
            return 'Jeanie Internet';
        case 'JNPB':
            return 'Jeanie Point of Bank';
        case 'JELM':
            return 'Jeanie LM';
        case 'JADV':
            return 'Jeanie Advtg';
        case 'JNIB':
            return 'Jeanie';
        case 'JPRB':
            return 'Jeanie Presto';
        case 'JPXB':
            return 'Exmpt Jeanie POS';
        case 'JLMB':
            return 'Jeanie LM';
        case 'JPBB':
            return 'Jeanie POB';
        case 'MJN2':
            return 'Jeanie RIC';
        case 'JADB':
            return 'Jeanie Advtg';
        case 'JP53':
            return 'Jeanie Preferred';
        case 'REVM':
            return 'Revolution Money';
        case 'CIRS':
            return 'Cirrus';
        case 'SHAZ':
            return 'Shazam';
        case 'PUGU':
            return 'Pulse-Gulfnet Reciprocal';
        case 'MAES':
            return 'Maestro';
        case 'PULS':
            return 'NCR Pulse';
        case 'AFN1':
            return 'Armed Forces Network';
        case 'NYGY':
            return 'NYCE Gateway';
        case 'ACCL':
            return 'Accel';
        case 'AKOP':
            return 'Alaska Option';
        case 'CTF1':
            return 'CU 24';
        case 'GDEB':
            return 'Generic Debit';
        case 'ITEL':
            return 'Instant Teller';
        case 'KETS':
            return 'Kansas Elect Trans';
        case 'MPCT':
            return 'MPACT';
        case 'TX00':
            return 'TX';
        case 'ECP':
            return 'Bankserv';
        case 'ECPG':
            return 'Total Check -Guarantee';
        case 'ECPT':
            return 'Total Check -Truncated';
        case 'ECPV':
            return 'Total Check Verification';
        case 'UNKN':
            return 'Unknown';
        case 'DBT1':
            return 'Debitman';
        case 'PINP':
            return 'Pin Prompting';
        case 'AFN2':
            return 'AFFN- 5/2 Switch';
        case 'AFMM':
            return 'AFFN MM';
        case 'AFDM':
            return 'AFFN DT MM';
        case 'AFPM':
            return 'AFFN Preferred Merchant';
        case 'MST1':
            return 'MAES Spmkt Tier 1';
        case 'MST2':
            return 'MAES Spmkt Tier 2';
        case 'MSBA':
            return 'MAES Spmkt Base';
        case 'MCT1':
            return 'MAES Conv Tier 1';
        case 'MCT2':
            return 'MAES Conv Tier 2';
        case 'MCBA':
            return 'MAES Conv Base';
        case 'MOT1':
            return 'MAES All Other Tier 1';
        case 'MOT2':
            return 'MAES All Other Tier 2';
        case 'MOBA':
            return 'MAES All Other Base';
        case 'AFFR':
            return 'AFFN Reversals';
        case 'AF2R':
            return 'AFN2 Reversals';
        case 'AML1':
            return 'AFFN/Maestro';
        case 'AFDP':
            return 'AFFN DT PM';
        case 'AML2':
            return 'AFFN/Maestro RCPRCLK';
        case 'JGT1':
            return 'Jeni T1 GR';
        case 'JGT2':
            return 'Jeni T2 GR';
        case 'JGT3':
            return 'Jeni T3 GR';
        case 'JOT1':
            return 'Jeni T1 OT';
        case 'JOT2':
            return 'Jeni T2 OT';
        case 'JOT3':
            return 'Jeni T3 OT';
        case 'JQSR':
            return 'JENI QSR';
        case 'JMED':
            return 'Jeni Med';
        case 'AJN1':
            return 'AFFN/ Jeanie 1';
        case 'AJN2':
            return 'AFFN/ Jeanie 2';
        case 'JAF1':
            return 'Jeanie /AFFN 1';
        case 'JAF2':
            return 'Jeanie / AFFN 2';
        case 'APR1':
            return 'AFN1 Presto!';
        case 'APR2':
            return 'AFN2 Presto!';
        case 'PDSD':
            return 'Pulse Discover Check Card';
        case 'BDSD':
            return 'Batch Discover';
        case 'PST1':
            return 'Presto!';
        case 'EBT':
            return 'EBT Transaction';
        case 'ENJ1':
            return 'New Jersey EBT';
        case 'ETX1':
            return 'Texas EBT';
        case 'EIL1':
            return 'Illinois EBT';
        case 'EOK1':
            return 'Oklahoma EBT';
        case 'EWCS':
            return 'AA Western Coalition EBT';
        case 'ENCS':
            return 'AA NE Coalition EBT';
        case 'ESAS':
            return 'AA Southern Alliance EBT';
        case 'ELA1':
            return 'Louisiana EBT';
        case 'EMD1':
            return 'Maryland EBT';
        case 'ESC1':
            return 'South Carolina EBT';
        case 'ENM1':
            return 'New Mexico EBT';
        case 'ETN1':
            return 'Tennessee EBT';
        case 'EUT1':
            return 'Utah EBT';
        case 'EQST':
            return 'AA-Quest Coalition EBT';
        case 'EDC1':
            return 'District of Columbia EBT';
        case 'EMI1':
            return 'Michigan EBT';
        case 'EIN1':
            return 'Indiana EBT';
        case 'EOR1':
            return 'Oregon EBT';
        case 'EDP1':
            return 'CA -San Diego EBT';
        case 'EDO1':
            return 'CA-San Bernardino EBT';
        case 'EVT1':
            return 'Vermont EBT';
        case 'ERI1':
            return 'Rhode Island EBT';
        case 'ENC1':
            return 'North Carolina EBT';
        case 'ENH1':
            return 'New Hampshire EBT';
        case 'EMN1':
            return 'Minnesota EBT';
        case 'EPA1':
            return 'Pennsylvania EBT';
        case 'EKS1':
            return 'Kansas EBT';
        case 'EWI1':
            return 'Wisconsin EBT';
        case 'ENS1':
            return 'North/South Dakota EBT';
        case 'EHI1':
            return 'Hawaii EBT';
        case 'EID1':
            return 'Idaho EBT';
        case 'EAK1':
            return 'Alaska EBT';
        case 'EWA1':
            return 'Washington EBT';
        case 'EAZ1':
            return 'Arizona EBT';
        case 'ECO1':
            return 'Colorado EBT';
        case 'EMA1':
            return 'Massachusetts EBT';
        case 'ECT1':
            return 'Connecticut EBT';
        case 'EME1':
            return 'Maine EBT';
        case 'ENY1':
            return 'New York EBT';
        case 'EAL1':
            return 'Alabama EBT';
        case 'EAR1':
            return 'Arkansas EBT';
        case 'EFL1':
            return 'Florida EBT';
        case 'EGA1':
            return 'Georgia EBT';
        case 'EKY1':
            return 'Kentucky EBT';
        case 'EMO1':
            return 'Missouri EBT';
        case 'EMS1':
            return 'Mississippi EBT';
        case 'EMT1':
            return 'Montana EBT';
        case 'ECA1':
            return 'California EBT';
        case 'EPR1':
            return 'Puerto Rico EBT';
        case 'EPR2':
            return 'Puerto Rico EBT';
        case 'ENE1':
            return 'Nebraska EBT';
        case 'EVA1':
            return 'Virginia EBT';
        case 'EPR3':
            return 'Puerto Rico EBT';
        case 'EDE1':
            return 'Delaware EBT';
        case 'ENV1':
            return 'Nevada EBT';
        case 'EBT':
            return 'Electronic Benefits Transfer';
        case 'EWV1':
            return 'West Virginia EBT';
        case 'EIA1':
            return 'Iowa EBT';
        case 'SZIA':
            return 'EBT Iowa Shazam';
        case 'EVI1':
            return 'Virgin Islands EBT';
        case 'EALL':
            return 'EBT All States';
        case 'EGU1':
            return 'Guam EBT';
        case 'PRFS':
            return 'EBT Puerto Rico - FS';
        case 'PRCS':
            return 'EBT Puerto Rico - CS';
        case 'EOH1':
            return 'Ohio EBT';
        case 'EWY1':
            return 'EBT Wyoming';
        case 'GIFT':
            return 'Gift Card Transaction';
        case 'VSCK':
            return 'POS Check Transaction';
        case 'VIS1':
            return 'Visa';
        case 'EHH1':
            return 'ECHO';
        case 'VPLN':
            return 'Visa ReadyLink';
        case 'BMLA':
            return 'Bill-Me-Later Core';
        case 'BMLB':
            return 'Bill-Me-Later 90 Days Same As Cash (SAC)';
        case 'BMLC':
            return 'Bill-Me-Later Business Core';
        case 'BMLD':
            return 'Bill-Me-Later Business 90 Days Same As Cash (SAC)';
        case 'BMLE':
            return 'Bill-Me-Later Private Label Core';
        case 'BMLF':
            return 'Bill-Me-Later Private Label 90 Days Same As Cash (SAC)';
        default:
            return 'Unknown';
    }
}

function getCardNetwork(network) {
    switch(network) {
        case "DISC":
        case "DISCV":
            return "DISCOVER"
        case "MCRD":
            return "MASTERCARD"
        case "VISA":
            return "VISA"
        case "AMEX":
            return "AMEX"
        default:
            return "OTHER"
    }
}

export {
    getCardNetworkType,
    getCardNetwork
}
